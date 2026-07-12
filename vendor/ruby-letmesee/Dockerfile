FROM httpd:2.4-bullseye

MAINTAINER kurema

# tzdata
RUN apt-get update -y && \
    apt-get install -y --no-install-recommends tzdata

ENV TZ=Asia/Tokyo

# dependencies
RUN apt-get install -y --no-install-recommends \
      ruby ruby-dev eb-utils libeb16-dev git build-essential ca-certificates \
      ispell ienglish-common iamerican-small && \
    apt-get clean && \
    rm -rf /var/cache/apt/archives/*

# enable CGI
RUN sed -ri 's/#LoadModule cgid_module/LoadModule cgid_module/g; \
             s/DirectoryIndex index.html/DirectoryIndex index.rb index.cgi index.html/g; \
             s/Options Indexes FollowSymLinks/Options Indexes FollowSymLinks ExecCGI/g; \
             s/#Scriptsock cgisock/Scriptsock cgisock/g; \
             s/#AddHandler cgi-script .cgi/AddHandler cgi-script .pl .rb .cgi/g' \
             /usr/local/apache2/conf/httpd.conf

RUN mkdir -p /usr/local/apache2/logs && touch /usr/local/apache2/logs/dummy

# build rubyeb19
RUN cd /tmp && mkdir src && cd src && \
    git clone https://github.com/kubo/rubyeb19.git && \
    cd rubyeb19 && \
    ruby extconf.rb && \
    make && make install

RUN apt remove -y git build-essential && apt autoremove -y

# copy letmesee
COPY . /usr/local/apache2/htdocs/

# fix shebang
RUN sed -i "1s|^#!.*|#!/usr/bin/ruby|" \
    /usr/local/apache2/htdocs/index.rb \
    /usr/local/apache2/htdocs/index.cgi 2>/dev/null || true

RUN chmod 755 /usr/local/apache2/htdocs/*.rb \
              /usr/local/apache2/htdocs/*.cgi 2>/dev/null || true

# startup
RUN echo "#!/bin/sh\nrm -f /usr/local/apache2/logs/cgisock.*\napachectl -D FOREGROUND" \
    > /startup.sh && chmod 755 /startup.sh

EXPOSE 80
CMD ["sh", "-c", "/startup.sh"]
