# ipatool-derived App Store module

This directory contains the runtime modules needed by AppleVault, derived from
`majd/ipatool` v2.5.0. The Cobra CLI, output formatting, and unrelated command
code are intentionally excluded.

AppleVault-specific changes include context-aware HTTP requests, explicit proxy
transport injection, and returning the remote IPA size with version metadata.

Copyright (c) 2021 majd

Licensed under the MIT License. The full license text is distributed at
`third_party/ipatool-LICENSE.txt`.
