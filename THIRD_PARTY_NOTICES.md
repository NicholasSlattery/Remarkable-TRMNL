# Third-party notices

This project interoperates with and was designed from the public formats of:

- AppLoad 0.5.3 (`asivery/rm-appload`), GPL-3.0. Its backend protocol and
  application manifest format are used for compatibility. The easy-installer
  release redistributes the unmodified ARM64 release binaries and a complete
  copy of the GPL-3.0 license. Corresponding source:
  `https://github.com/asivery/rm-appload/tree/v0.5.3`.
- XOVI 0.3.3 (`asivery/xovi`), LGPL-3.0, and rm-xovi-extensions release 19
  (`asivery/rm-xovi-extensions`), GPL-3.0. The easy-installer release
  redistributes the unmodified, checksum-pinned ARM64 runtime archive and copies
  the applicable license to the tablet. Corresponding source:
  `https://github.com/asivery/xovi/tree/v0.3.3` and
  `https://github.com/asivery/rm-xovi-extensions/tree/v19-23052026`.

The Windows easy-installer ZIP also contains the complete corresponding source
archives for all three redistributed runtime components in `Corresponding
Source/`, so the exact source remains available with the binaries.
- TRMNL Display (`usetrmnl/trmnl-display`), MIT. The Device API request shape,
  configuration concepts, and TRMNL icon were adapted. The upstream license is
  reproduced below.

The runtime components are separate programs. They are not linked into the MIT
licensed TRMNL backend, and the standalone application bundle contains none of
their code. Release builders can reproduce the runtime download using
`scripts/fetch-runtime.ps1`; its URLs and SHA-256 values are pinned.

## TRMNL Display MIT license

Copyright (c) 2025 TRMNL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
