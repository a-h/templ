# Code signing

Binaries are created by the GitHub Actions workflow at https://github.com/a-h/templ/blob/main/.github/workflows/release.yml

Binaries and Docker images are signed by cosign. The public key is stored in the repository at https://github.com/a-h/templ/blob/main/cosign.pub

Each release includes a `checksums.txt` file listing the SHA-256 checksum of each binary archive, and a `checksums.txt.sigstore.json` file containing the signature of `checksums.txt`.

To verify the checksums file:

```bash
cosign verify-blob --key cosign.pub --bundle checksums.txt.sigstore.json checksums.txt
```

To verify a Docker image:

```bash
cosign verify --key cosign.pub ghcr.io/a-h/templ:latest
```

Each Docker image has a signed SPDX SBOM attached as an attestation. To verify it and print the SBOM for the `linux/amd64` image:

```bash
cosign verify-attestation --key cosign.pub --type spdxjson \
  "ghcr.io/a-h/templ@$(crane digest --platform linux/amd64 ghcr.io/a-h/templ:latest)" \
  | jq -r '.payload | @base64d | fromjson | .predicate'
```

Instructions for key verification at https://docs.sigstore.dev/verifying/verify/
