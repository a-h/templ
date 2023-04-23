FROM nixos/nix
RUN nix-channel --update
COPY ./nix/nix.conf /etc/nix/nix.conf
RUN mkdir /templ
COPY . /templ
WORKDIR /templ
RUN nix develop --impure --command printf "Build complete\n"
COPY ./nix/.config /root/.config
# Open port for templ LSP HTTP debugging
EXPOSE 7474
CMD nix develop --impure
