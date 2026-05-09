# Copyright 2026 Optiqor contributors
# SPDX-License-Identifier: Apache-2.0
#
# vhs tape — records the kerno doctor MVP demo for the README hero GIF.
#
# IMPORTANT: vhs records a non-interactive shell that cannot type a sudo
# password. Either run vhs itself as root (Makefile does this) or
# pre-grant CAP_BPF to ./bin/kerno with `setcap` so no sudo is needed.
#
# Run with:    make demo
# Direct:      sudo $(which vhs) demo.tape
# Output:      demo.gif (~5 MB target after gifsicle)

Output demo.gif

# ── visual config ──
Set Shell      "bash"
Set FontSize   14
Set Width      1100
Set Height     720
Set Theme      "Dracula"
Set TypingSpeed 35ms
Set Padding    20

# ── intro ──
Type "# Kerno doctor — kernel-level incident diagnosis in one command"
Sleep 1s
Enter
Sleep 500ms

Type "# Step 1: induce a synthetic disk-fsync bottleneck"
Sleep 600ms
Enter
Type "./bin/kerno chaos --induce disk-sat --duration 25s --intensity high --yes &"
Enter
Sleep 2s

Type "clear"
Enter
Sleep 500ms

# ── doctor run ── (recorded shell is root via 'sudo vhs ...', so no
# 'sudo' needed inside — that's what unblocks the password prompt.)
Type "# Step 2: ask kerno what's wrong"
Sleep 600ms
Enter
Type "./bin/kerno --config scripts/verify-config.yaml doctor --duration 10s"
Enter

# Doctor takes ~10s to collect + a couple seconds to render the
# box-drawn report. Wait for output to fully paint.
Sleep 15s

# ── outro ──
Type "# Root cause identified. Same flow runs in CI: 'make verify'."
Sleep 3s
