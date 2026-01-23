created: 2026-01-22T17:57
title: Add email verification to waitlist flow
area: api
files:
  - src/pages/api/waitlist.ts

## Problem

The current waitlist signup flow accepts any email without verification. This leads to:
- Low quality email addresses (typos, fake emails)
- Potential spam signups
- Reduced deliverability when sending launch announcements
- Wasted effort on invalid leads

Need to verify emails are real and owned by the person signing up before adding to waitlist.

## Solution

Options to consider:
1. **Double opt-in**: Send verification email with confirmation link before adding to waitlist
2. **Email validation API**: Use service like ZeroBounce, Hunter.io, or Neverbounce to validate email format and deliverability
3. **Combination**: Basic validation + verification email for high-value confirmation

TBD - need to decide on approach based on friction tolerance vs quality tradeoff.
