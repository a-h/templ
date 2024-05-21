# Code signing

Binaries are created by the Github Actions workflow at https://github.com/a-h/templ/blob/main/.github/workflows/release.yml

Binaries are signed by cosign. The public key is stored in the repository at https://github.com/a-h/templ/blob/main/cosign.pub

Instructions for key verification at https://docs.sigstore.dev/verifying/verify/
