package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/asticode/go-astiav"
)

// webmCache stores transcoded WebM output keyed by SHA-256 of the raw input.
// Dictionary video clips are short and repeated requests are common, so caching
// avoids re-encoding the same clip on every request.
var webmCache sync.Map

// TranscodeToWebM converts a raw MPEG-1 video elementary stream to WebM (VP8).
// It is safe to call concurrently; results are deduplicated via the cache.
func TranscodeToWebM(mpeg1Data []byte) ([]byte, error) {
	key := sha256.Sum256(mpeg1Data)
	if v, ok := webmCache.Load(key); ok {
		return v.([]byte), nil
	}

	webm, err := transcodeOnce(mpeg1Data)
	if err != nil {
		return nil, fmt.Errorf("transcode mpeg1→webm: %w", err)
	}
	webmCache.Store(key, webm)
	return webm, nil
}

// transcodeOnce writes the input to a temp file (so libavformat can probe the
// container format), transcodes to a second temp file, reads the result into
// memory, and cleans up both temp files.
func transcodeOnce(input []byte) ([]byte, error) {
	inFile, err := os.CreateTemp("", "lms-in-*.mpg")
	if err != nil {
		return nil, fmt.Errorf("create input temp file: %w", err)
	}
	inPath := inFile.Name()
	defer os.Remove(inPath)
	if _, err := inFile.Write(input); err != nil {
		inFile.Close()
		return nil, fmt.Errorf("write input temp file: %w", err)
	}
	inFile.Close()

	outFile, err := os.CreateTemp("", "lms-out-*.webm")
	if err != nil {
		return nil, fmt.Errorf("create output temp file: %w", err)
	}
	outPath := outFile.Name()
	outFile.Close()
	defer os.Remove(outPath)

	if err := transcodeFile(inPath, outPath); err != nil {
		return nil, err
	}
	return os.ReadFile(outPath)
}

// transcodeFile transcodes the video at inPath to a WebM (VP8) file at outPath.
func transcodeFile(inPath, outPath string) error {
	astiav.SetLogLevel(astiav.LogLevelError)

	// ------------------------------------------------------------------ //
	// Input                                                                //
	// ------------------------------------------------------------------ //

	inFmtCtx := astiav.AllocFormatContext()
	if inFmtCtx == nil {
		return fmt.Errorf("alloc input format context")
	}
	defer inFmtCtx.CloseInput()

	if err := inFmtCtx.OpenInput(inPath, nil, nil); err != nil {
		return fmt.Errorf("open input %q: %w", inPath, err)
	}
	if err := inFmtCtx.FindStreamInfo(nil); err != nil {
		return fmt.Errorf("find stream info: %w", err)
	}

	var videoStream *astiav.Stream
	for _, s := range inFmtCtx.Streams() {
		if s.CodecParameters().MediaType() == astiav.MediaTypeVideo {
			videoStream = s
			break
		}
	}
	if videoStream == nil {
		return fmt.Errorf("no video stream found in input")
	}

	decoder := astiav.FindDecoder(videoStream.CodecParameters().CodecID())
	if decoder == nil {
		return fmt.Errorf("no decoder for codec %v", videoStream.CodecParameters().CodecID())
	}
	decCtx := astiav.AllocCodecContext(decoder)
	if decCtx == nil {
		return fmt.Errorf("alloc decoder context")
	}
	defer decCtx.Free()

	if err := videoStream.CodecParameters().ToCodecContext(decCtx); err != nil {
		return fmt.Errorf("copy codec params to decoder context: %w", err)
	}
	decCtx.SetThreadCount(0) // let FFmpeg choose the thread count
	if err := decCtx.Open(decoder, nil); err != nil {
		return fmt.Errorf("open decoder: %w", err)
	}

	// Derive the per-frame duration in milliseconds from the stream frame rate.
	// Fall back to 25 fps (40 ms/frame) when the stream carries no rate info.
	frameDurMs := int64(40)
	if fps := videoStream.AvgFrameRate(); fps.Num() > 0 && fps.Den() > 0 {
		frameDurMs = int64(1000) * int64(fps.Den()) / int64(fps.Num())
		if frameDurMs <= 0 {
			frameDurMs = 40
		}
	}

	// ------------------------------------------------------------------ //
	// Output                                                               //
	// ------------------------------------------------------------------ //

	outFmtCtx, err := astiav.AllocOutputFormatContext(nil, "webm", outPath)
	if err != nil {
		return fmt.Errorf("alloc output format context: %w", err)
	}
	defer outFmtCtx.Free()

	// Prefer VP8 (libvpx); fall back to VP9 (libvpx-vp9).
	encoder := astiav.FindEncoderByName("libvpx")
	if encoder == nil {
		encoder = astiav.FindEncoderByName("libvpx-vp9")
	}
	if encoder == nil {
		return fmt.Errorf("VP8/VP9 encoder not available; install libvpx")
	}

	outStream := outFmtCtx.NewStream(encoder)
	if outStream == nil {
		return fmt.Errorf("new output stream")
	}

	encCtx := astiav.AllocCodecContext(encoder)
	if encCtx == nil {
		return fmt.Errorf("alloc encoder context")
	}
	defer encCtx.Free()

	encCtx.SetWidth(decCtx.Width())
	encCtx.SetHeight(decCtx.Height())
	encCtx.SetSampleAspectRatio(decCtx.SampleAspectRatio())
	encCtx.SetPixelFormat(astiav.PixelFormatYuv420P)
	encCtx.SetTimeBase(astiav.NewRational(1, 1000)) // millisecond timebase
	// CRF mode: b:v 0 disables bitrate targeting; crf drives quality (0=best, 63=worst).
	encCtx.SetBitRate(0)

	if outFmtCtx.OutputFormat().Flags().Has(astiav.IOFormatFlagGlobalheader) {
		encCtx.SetFlags(encCtx.Flags().Add(astiav.CodecContextFlagGlobalHeader))
	}

	// Enable CRF quality mode for libvpx.
	encDict := astiav.NewDictionary()
	defer encDict.Free()
	encDict.Set("crf", "10", astiav.NewDictionaryFlags())
	encDict.Set("b", "0", astiav.NewDictionaryFlags())

	if err := encCtx.Open(encoder, encDict); err != nil {
		return fmt.Errorf("open encoder: %w", err)
	}
	if err := encCtx.ToCodecParameters(outStream.CodecParameters()); err != nil {
		return fmt.Errorf("copy encoder params to output stream: %w", err)
	}
	outStream.SetTimeBase(encCtx.TimeBase())

	// Open the output file for writing.
	if !outFmtCtx.OutputFormat().Flags().Has(astiav.IOFormatFlagNofile) {
		pb, err := astiav.OpenIOContext(outPath, astiav.NewIOContextFlags(astiav.IOContextFlagWrite), nil, nil)
		if err != nil {
			return fmt.Errorf("open output IO context: %w", err)
		}
		defer pb.Free()
		outFmtCtx.SetPb(pb)
	}

	if err := outFmtCtx.WriteHeader(nil); err != nil {
		return fmt.Errorf("write WebM header: %w", err)
	}

	// ------------------------------------------------------------------ //
	// Transcode loop                                                        //
	// ------------------------------------------------------------------ //

	pkt := astiav.AllocPacket()
	defer pkt.Free()
	decFrm := astiav.AllocFrame()
	defer decFrm.Free()
	encFrm := astiav.AllocFrame()
	defer encFrm.Free()

	var swsCtx *astiav.SoftwareScaleContext
	defer func() {
		if swsCtx != nil {
			swsCtx.Free()
		}
	}()

	// frameCount drives PTS calculation.  Each frame's PTS is
	// frameCount * frameDurMs, giving correct playback speed regardless
	// of what timestamps the source stream carries.
	var frameCount int64

	// sendEncode sends a frame (or nil to flush) to the encoder and writes all
	// resulting packets to the output.
	sendEncode := func(frame *astiav.Frame) error {
		if err := encCtx.SendFrame(frame); err != nil {
			return err
		}
		for {
			if err := encCtx.ReceivePacket(pkt); err != nil {
				if errors.Is(err, astiav.ErrEagain) || astiav.ErrEof.Is(err) {
					break
				}
				return err
			}
			pkt.SetStreamIndex(outStream.Index())
			pkt.RescaleTs(encCtx.TimeBase(), outStream.TimeBase())
			if err := outFmtCtx.WriteInterleavedFrame(pkt); err != nil {
				pkt.Unref()
				return fmt.Errorf("write interleaved frame: %w", err)
			}
			pkt.Unref()
		}
		return nil
	}

	// processFrame converts a decoded frame to YUV420P if needed, then encodes it.
	processFrame := func(decFrm *astiav.Frame) error {
		srcFmt := decFrm.PixelFormat()
		var sendFrm *astiav.Frame

		pts := frameCount * frameDurMs
		frameCount++

		if srcFmt == astiav.PixelFormatYuv420P {
			// No pixel format conversion needed.
			if err := encFrm.Ref(decFrm); err != nil {
				return nil // non-fatal; skip this frame
			}
			encFrm.SetPts(pts)
			sendFrm = encFrm
		} else {
			// Lazy-init the swscale context on the first frame that needs conversion.
			if swsCtx == nil {
				var err error
				swsCtx, err = astiav.CreateSoftwareScaleContext(
					decFrm.Width(), decFrm.Height(), srcFmt,
					decFrm.Width(), decFrm.Height(), astiav.PixelFormatYuv420P,
					astiav.NewSoftwareScaleContextFlags(astiav.SoftwareScaleContextFlagBilinear),
				)
				if err != nil {
					return fmt.Errorf("create sws context: %w", err)
				}
				encFrm.SetWidth(decFrm.Width())
				encFrm.SetHeight(decFrm.Height())
				encFrm.SetPixelFormat(astiav.PixelFormatYuv420P)
				if err := encFrm.AllocBuffer(0); err != nil {
					return fmt.Errorf("alloc encode frame buffer: %w", err)
				}
			}
			if err := swsCtx.ScaleFrame(decFrm, encFrm); err != nil {
				return nil // non-fatal; skip this frame
			}
			encFrm.SetPts(pts)
			sendFrm = encFrm
		}

		err := sendEncode(sendFrm)
		encFrm.Unref()
		return err
	}

	// receiveFrames drains all decoded frames from the decoder.
	receiveFrames := func() error {
		for {
			if err := decCtx.ReceiveFrame(decFrm); err != nil {
				if errors.Is(err, astiav.ErrEagain) || astiav.ErrEof.Is(err) {
					return nil
				}
				return err
			}
			if err := processFrame(decFrm); err != nil {
				decFrm.Unref()
				return err
			}
			decFrm.Unref()
		}
	}

	for {
		if err := inFmtCtx.ReadFrame(pkt); err != nil {
			if astiav.ErrEof.Is(err) {
				break
			}
			pkt.Unref()
			continue
		}

		if pkt.StreamIndex() != videoStream.Index() {
			pkt.Unref()
			continue
		}

		if err := decCtx.SendPacket(pkt); err != nil {
			pkt.Unref()
			continue
		}
		pkt.Unref()

		if err := receiveFrames(); err != nil {
			return err
		}
	}

	// Flush the decoder, then the encoder.
	_ = decCtx.SendPacket(nil)
	_ = receiveFrames()
	_ = sendEncode(nil)

	return outFmtCtx.WriteTrailer()
}
