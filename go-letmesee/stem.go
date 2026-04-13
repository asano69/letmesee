package main

import "strings"

// irregularVerbs maps inflected forms to their base (infinitive) form.
// Ported directly from stem.rb.
var irregularVerbs = map[string]string{
	"arisen": "arise", "arose": "arise", "arise": "arise",
	"awoken": "awake", "awakened": "awake", "awoke": "awake", "awake": "awake",
	"been": "be", "were": "be", "was": "be", "be": "be",
	"borne": "bear", "born": "bear", "bore": "bear", "bear": "bear",
	"beat": "beat", "beaten": "beat",
	"become": "become", "became": "become",
	"befallen": "befall", "befell": "befall", "befall": "befall",
	"begun": "begin", "began": "begin", "begin": "begin",
	"beheld": "behold", "behold": "behold",
	"bent": "bend", "bind": "bind", "bound": "bind",
	"bitten": "bite", "bit": "bite", "bite": "bite",
	"bled": "bleed", "bleed": "bleed",
	"blown": "blow", "blew": "blow", "blow": "blow",
	"broken": "break", "broke": "break", "break": "break",
	"bred": "breed", "breed": "breed",
	"brought": "bring", "bring": "bring",
	"broadcast": "broadcast",
	"built":     "build", "build": "build",
	"burned": "burn", "burnt": "burn", "burn": "burn",
	"burst":  "burst",
	"busted": "bust", "bust": "bust",
	"bought": "buy", "buy": "buy",
	"cast":   "cast",
	"caught": "catch", "catch": "catch",
	"chosen": "choose", "chose": "choose", "choose": "choose",
	"clung": "cling", "cling": "cling",
	"come": "come", "came": "come",
	"cost":  "cost",
	"crept": "creep", "creep": "creep",
	"cut":   "cut",
	"dealt": "deal", "deal": "deal",
	"dug": "dig", "dig": "dig",
	"dived": "dive", "dove": "dive", "dive": "dive",
	"done": "do", "did": "do", "do": "do",
	"drawn": "draw", "drew": "draw", "draw": "draw",
	"dreamed": "dream", "dreamt": "dream", "dream": "dream",
	"drunk": "drink", "drank": "drink", "drink": "drink",
	"driven": "drive", "drove": "drive", "drive": "drive",
	"dwelled": "dwell", "dwelt": "dwell", "dwell": "dwell",
	"eaten": "eat", "ate": "eat", "eat": "eat",
	"fallen": "fall", "fell": "fall", "fall": "fall",
	"fed": "feed", "feed": "feed",
	"felt": "feel", "feel": "feel",
	"fought": "fight", "fight": "fight",
	"found": "find", "find": "find",
	"fit": "fit", "fitted": "fit",
	"fled": "flee", "flee": "flee",
	"flung": "fling", "fling": "fling",
	"flown": "fly", "flew": "fly", "fly": "fly",
	"forbidden": "forbid", "forbade": "forbid", "forbid": "forbid",
	"forecast":  "forecast",
	"forgotten": "forget", "forgot": "forget", "forget": "forget",
	"forgiven": "forgive", "forgave": "forgive", "forgive": "forgive",
	"forsaken": "forsake", "forsook": "forsake", "forsake": "forsake",
	"frozen": "freeze", "froze": "freeze", "freeze": "freeze",
	"got": "get", "gotten": "get", "get": "get",
	"given": "give", "gave": "give", "give": "give",
	"gone": "go", "went": "go", "go": "go",
	"ground": "grind", "grind": "grind",
	"grown": "grow", "grew": "grow", "grow": "grow",
	"hung": "hang", "hang": "hang",
	"had": "have", "have": "have", "has": "have",
	"heard": "hear", "hear": "hear",
	"hidden": "hide", "hid": "hide", "hide": "hide",
	"hit":  "hit",
	"held": "hold", "hold": "hold",
	"hurt": "hurt",
	"kept": "keep", "keep": "keep",
	"kneeled": "kneel", "knelt": "kneel", "kneel": "kneel",
	"knitted": "knit", "knit": "knit",
	"known": "know", "knew": "know", "know": "know",
	"laid": "lay", "lay": "lay",
	"led": "lead", "lead": "lead",
	"leant": "lean", "leaned": "lean", "lean": "lean",
	"leaped": "leap", "leapt": "leap", "leap": "leap",
	"learnt": "learn", "learned": "learn", "learn": "learn",
	"left": "leave", "leave": "leave",
	"lent": "lend", "lend": "lend",
	"let":  "let",
	"lain": "lie", "lie": "lie",
	"lighted": "light", "lit": "light", "light": "light",
	"lost": "lose", "lose": "lose",
	"made": "make", "make": "make",
	"meant": "mean", "mean": "mean",
	"met": "meet", "meet": "meet",
	"mislaid": "mislay", "mislay": "mislay",
	"misled": "mislead", "mislead": "mislead",
	"misread":  "misread",
	"misspelt": "misspell", "misspelled": "misspell", "misspell": "misspell",
	"mistaken": "mistake", "mistook": "mistake", "mistake": "mistake",
	"misunderstood": "misunderstand", "misunderstand": "misunderstand",
	"overcome": "overcome", "overcame": "overcome",
	"overdone": "overdo", "overdid": "overdo", "overdo": "overdo",
	"overrun": "overrun", "overran": "overrun",
	"overseen": "oversee", "oversaw": "oversee", "oversee": "oversee",
	"overthrown": "overthrow", "overthrew": "overthrow", "overthrow": "overthrow",
	"paid": "pay", "pay": "pay",
	"put":     "put",
	"quit":    "quit",
	"read":    "read",
	"rebuilt": "rebuild", "rebuild": "rebuild",
	"redo": "redo", "redid": "redo",
	"rerun": "rerun", "reran": "rerun",
	"reset":     "reset",
	"rewritten": "rewrite", "rewrote": "rewrite", "rewrite": "rewrite",
	"rid":    "rid",
	"ridden": "ride", "rode": "ride", "ride": "ride",
	"rung": "ring", "rang": "ring", "ring": "ring",
	"risen": "rise", "rose": "rise", "rise": "rise",
	"run": "run", "ran": "run",
	"said": "say", "say": "say",
	"seen": "see", "saw": "see", "see": "see",
	"sought": "seek", "seek": "seek",
	"sold": "sell", "sell": "sell",
	"sent": "send", "send": "send",
	"set":    "set",
	"shaken": "shake", "shook": "shake", "shake": "shake",
	"shed":  "shed",
	"shone": "shine", "shined": "shine", "shine": "shine",
	"shot": "shoot", "shoot": "shoot",
	"showed": "show", "shown": "show", "show": "show",
	"shrunk": "shrink", "shrank": "shrink", "shrink": "shrink",
	"shut": "shut",
	"sung": "sing", "sang": "sing", "sing": "sing",
	"sat": "sit", "sit": "sit",
	"slept": "sleep", "sleep": "sleep",
	"slid": "slide", "slide": "slide",
	"slit":   "slit",
	"spoken": "speak", "spoke": "speak", "speak": "speak",
	"sped": "speed", "speed": "speed",
	"spelt": "spell", "spelled": "spell", "spell": "spell",
	"spent": "spend", "spend": "spend",
	"spun": "spin", "spin": "spin",
	"split":  "split",
	"spread": "spread",
	"sprung": "spring", "sprang": "spring", "spring": "spring",
	"stood": "stand", "stand": "stand",
	"stolen": "steal", "stole": "steal", "steal": "steal",
	"stuck": "stick", "stick": "stick",
	"stung": "sting", "sting": "sting",
	"stunk": "stink", "stank": "stink", "stink": "stink",
	"stridden": "stride", "strode": "stride", "stride": "stride",
	"striven": "strive", "strove": "strive", "strive": "strive",
	"struck": "strike", "strike": "strike",
	"strung": "string", "string": "string",
	"sworn": "swear", "swore": "swear", "swear": "swear",
	"swept": "sweep", "sweep": "sweep",
	"swum": "swim", "swam": "swim", "swim": "swim",
	"swung": "swing", "swing": "swing",
	"taken": "take", "took": "take", "take": "take",
	"taught": "teach", "teach": "teach",
	"torn": "tear", "tore": "tear", "tear": "tear",
	"told": "tell", "tell": "tell",
	"thought": "think", "think": "think",
	"thrown": "throw", "threw": "throw", "throw": "throw",
	"thrust": "thrust",
	"trod":   "tread", "trodden": "tread", "tread": "tread",
	"understood": "understand", "understand": "understand",
	"undertaken": "undertake", "undertook": "undertake", "undertake": "undertake",
	"underwritten": "underwrite", "underwrote": "underwrite", "underwrite": "underwrite",
	"undone": "undo", "undid": "undo", "undo": "undo",
	"upheld": "uphold", "uphold": "uphold",
	"upset": "upset",
	"woken": "wake", "woke": "wake", "wake": "wake",
	"worn": "wear", "wore": "wear", "wear": "wear",
	"woven": "weave", "wove": "weave", "weave": "weave",
	"wept": "weep", "weep": "weep",
	"won": "win", "win": "win",
	"wound": "wind", "wind": "wind",
	"withdrawn": "withdraw", "withdrew": "withdraw", "withdraw": "withdraw",
	"wrung": "wring", "wring": "wring",
	"written": "write", "wrote": "write", "write": "write",
	// Auxiliaries and pronouns
	"is": "be", "are": "be", "am": "be",
	"does": "do",
	"may":  "may", "might": "may",
	"can": "can", "could": "can",
	"will": "will", "would": "will",
	"must": "must",
	"good": "good", "better": "good", "best": "good",
	"bad": "bad", "worse": "bad", "worst": "bad",
	"little": "little", "less": "little", "least": "little",
	"far": "far", "further": "far", "furthest": "far",
	"farther": "far", "farthest": "far",
	"old": "old", "elder": "old", "eldest": "old",
	"a": "a", "an": "a",
	"my": "one's", "your": "one's", "his": "one's", "her": "one's",
	"its": "one's", "our": "one's", "their": "one's",
	"myself": "oneself", "yourself": "oneself", "himself": "oneself",
	"herself": "oneself", "itself": "oneself", "ourselves": "oneself",
	"yourselves": "oneself", "themselves": "oneself",
}

// isVowel reports whether b is an ASCII vowel (including 'y').
func isVowel(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u', 'y':
		return true
	}
	return false
}

// hasVowel reports whether s contains at least one vowel.
func hasVowel(s string) bool {
	for i := 0; i < len(s); i++ {
		if isVowel(s[i]) {
			return true
		}
	}
	return false
}

// hasCVC returns true when s matches the pattern consonant*vowel*consonant
// (i.e. ends with at least one C after a V after optional leading Cs).
// This mirrors the Ruby M regexp check used in stemImpl.
func hasCVC(s string) bool {
	// We need at least one vowel cluster in s.
	return hasVowel(s)
}

// stemImpl returns all candidate base forms for a single lowercase word.
// It is a Go port of Stem#stem_impl from stem.rb.
func stemImpl(word string) []string {
	if len(word) < 3 {
		return []string{word}
	}
	// Only ASCII words.
	for _, c := range word {
		if c > 127 {
			return []string{word}
		}
	}
	// Apostrophe: strip suffix.
	if idx := strings.IndexByte(word, '\''); idx >= 0 {
		return []string{word, word[:idx]}
	}

	// -s endings
	if strings.HasSuffix(word, "s") {
		stem := word[:len(word)-1]
		if strings.HasSuffix(word, "sses") || strings.HasSuffix(word, "shes") {
			return []string{word, word[:len(word)-2]}
		}
		if strings.HasSuffix(word, "ies") {
			return []string{word, stem[:len(stem)-1] + "y", stem}
		}
		if strings.HasSuffix(word, "sses") {
			return []string{word, word[:len(word)-2]}
		}
		if len(stem) > 0 && stem[len(stem)-1] != 's' {
			return []string{word, stem}
		}
		return []string{word}
	}

	// -ed / -ing endings
	if strings.HasSuffix(word, "ed") || strings.HasSuffix(word, "ing") {
		var stem string
		if strings.HasSuffix(word, "ed") {
			stem = word[:len(word)-2]
		} else {
			stem = word[:len(word)-3]
		}
		if strings.HasSuffix(word, "ied") {
			return []string{word, stem + "y", stem}
		}
		if strings.HasSuffix(word, "eed") {
			if hasCVC(word[:len(word)-3]) {
				return []string{word, word[:len(word)-1]}
			}
			return []string{word}
		}
		if hasCVC(stem) {
			if strings.HasSuffix(stem, "at") || strings.HasSuffix(stem, "bl") || strings.HasSuffix(stem, "v") {
				return []string{word, stem + "e"}
			}
			if len(stem) > 1 && stem[len(stem)-1] == stem[len(stem)-2] && !isVowel(stem[len(stem)-1]) {
				shorter := stem[:len(stem)-1]
				if len(stem) > 3 {
					return []string{word, stem, shorter}
				}
				return []string{word, stem}
			}
			return []string{word, stem + "e", stem}
		}
	}

	// -er / -est endings
	if strings.HasSuffix(word, "er") || strings.HasSuffix(word, "est") {
		var stem string
		if strings.HasSuffix(word, "er") {
			stem = word[:len(word)-2]
		} else {
			stem = word[:len(word)-3]
		}
		if strings.HasSuffix(word, "ier") || strings.HasSuffix(word, "iest") {
			return []string{word, stem + "y"}
		}
		if hasCVC(stem) {
			if strings.HasSuffix(stem, "at") || strings.HasSuffix(stem, "bl") ||
				strings.HasSuffix(stem, "iz") || strings.HasSuffix(stem, "nc") ||
				strings.HasSuffix(stem, "v") {
				return []string{word, stem + "e"}
			}
			if len(stem) > 1 && stem[len(stem)-1] == stem[len(stem)-2] && !isVowel(stem[len(stem)-1]) {
				shorter := stem[:len(stem)-1]
				if len(stem) > 3 {
					return []string{word, stem, shorter}
				}
				return []string{word, stem}
			}
			return []string{word, stem + "e", stem}
		}
	}

	return []string{word}
}

// Stem returns a pipe-separated string of candidate forms for word,
// mirroring the Ruby Stem.stem() function used when querying dictionaries.
func Stem(word string) string {
	lower := strings.ToLower(word)
	if base, ok := irregularVerbs[lower]; ok {
		if base == lower {
			return lower
		}
		return lower + "|" + base
	}
	forms := stemImpl(lower)
	return strings.Join(dedupSlice(forms), "|")
}

func dedupSlice(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
