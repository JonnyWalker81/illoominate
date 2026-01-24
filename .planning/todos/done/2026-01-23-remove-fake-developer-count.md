---
created: 2026-01-23T22:01
title: Remove fake developer count from TrustBlock
area: ui
files:
  - src/components/TrustBlock.astro:43
---

## Problem

The TrustBlock component displays "Join 142 developers already on the waitlist" but there are no actual users yet. This is misleading and should be removed until there's real traction to display.

## Solution

Remove or replace the line at `src/components/TrustBlock.astro:43`. Options:
- Remove the line entirely
- Replace with a non-numeric trust signal (e.g., "Be among the first to try")
- Make it dynamic based on actual waitlist count (if > threshold)
