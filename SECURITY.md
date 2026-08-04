# Security policy

## Supported version

Security fixes are provided for the newest release only.

## Reporting a vulnerability

Do not open a public issue for credential exposure, installer command execution,
path traversal, host-key verification, or device recovery vulnerabilities. Use
GitHub's private vulnerability reporting feature for this repository. If that
feature is unavailable, contact the repository owner privately and include the
affected version, reproduction steps, and impact.

Do not include a real TRMNL API key, tablet SSH password, private key, device IP,
or unredacted configuration in a report.

## Security design

- The graphical installer listens only on a random `127.0.0.1` port, uses a
  per-run CSRF token, and requires confirmation of the tablet SSH fingerprint.
- Payload files and downloaded upstream runtimes are SHA-256 verified.
- Remote Device API/BYOS connections require HTTPS.
- Settings are stored owner-only and diagnostics redact the API key.
- Runtime installation rolls back newly installed files on failure.
- API redirects cannot cross origin/protocol, remote images require HTTPS, and
  downloaded images are size/dimension validated before decoding.
- Tag releases run the CI validation matrix before packaging, publish SHA-256
  checksums and an SBOM, and support Authenticode signing. Public repository
  releases also publish GitHub build-provenance attestations.

Checksum verification establishes that a file matches the GitHub Release; it
does not identify a Windows publisher. Users should expect **Unknown publisher**
until the release notes explicitly say the installer was Authenticode-signed and
the signature verifies to the named maintainer.
