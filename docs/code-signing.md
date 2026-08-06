# Self-signed Windows release

This community build uses an Authenticode signature made with a self-signed
code-signing certificate. The signature protects the executable against changes
after signing, but the certificate is **not publicly trusted**. Windows may still
show Unknown Publisher, SmartScreen, or an untrusted-certificate warning.

The release archive includes:

- `SELF-SIGNED-CERTIFICATE.cer`: the public certificate only; it contains no
  private signing key.
- `SELF-SIGNED-CERTIFICATE.txt`: its subject, validity, thumbprint, and SHA-256.

Verify the release ZIP against the separately published `SHA256SUMS.txt` and,
when available, the GitHub artifact attestation. The certificate inside the ZIP
does not independently prove who published the ZIP. Do not install it into a
trusted certificate store unless you have independently verified its fingerprint
and accept that trust decision.

Maintainers create this build with:

```powershell
./scripts/build-release.ps1 -Version 2.0.0
./scripts/sign-self-signed-release.ps1 -Version 2.0.0
```

The private key remains in the maintainer's Windows Current User certificate
store for reuse and is never added to the repository or release archive.
