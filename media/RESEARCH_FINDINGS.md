# Deep research findings — annotated

Result of running [DEEP_RESEARCH_PROMPT.md](DEEP_RESEARCH_PROMPT.md), with our
corrections. Run date: 2026-08-05. The researcher had no access to this repository by
design, so it evaluated the *concept* of TRMNL-on-reMarkable, not this implementation.
That's what makes the verdicts worth keeping and several of the objections wrong.

## Their verdicts

- **Kinship: superficial.** Shared e-ink and a shared minimalist marketing register, but
  opposed design positions — TRMNL pushes a scheduled sequence into your periphery,
  reMarkable waits for you to pull. Calm-tech mapping: TRMNL wins on peripheral attention
  and "least technology needed"; reMarkable wins on respecting social context and on not
  interrupting, but cannot inform peripherally at all.
- **Usefulness: a clever hack for a narrow niche**, not a broadly useful product.
- **Audience: low hundreds to ~2,000** worldwide.

## Accept

1. **Drop any claim of shared philosophy.** The honest line is shared screen technology
   plus idle hardware. Applied to `PITCH.md` §10.
2. **Battery is the top practical risk**, ahead of Developer Mode. A dedicated dashboard
   idles for months because its firmware does nothing between refreshes; a tablet with a
   full OS can't match that. Applied to `PITCH.md` §10 with interval guidance.
3. **Audience is small and should be planned for as small.** Applied to `PITCH.md` §7.
4. **"If you don't already own a Paper Pro, buy a dedicated dashboard."** Saying this
   costs nothing and buys credibility. Applied to `PITCH.md` §7 and §10.
5. **Paper-Pro-specific layouts are the one real product gap** — templates built for
   1620×2160, high contrast, few grey levels. Everything else they recommended is shipped.

## Reject, with reasons

| Their claim | Why it's wrong |
|---|---|
| "Can't take notes once it's a dashboard; you've lost the tablet's main function" | **Factually wrong.** AppLoad coexists with the stock interface — swipe, button, or upper-left hold returns to the notebook, and a reboot lands on stock. This was their strongest objection and it doesn't hold. |
| "No official port; you'd need to write a daemon; any OS upgrade breaks your script" | Describes the absence of this project. The app, installer, firmware gating, and recovery paths exist. |
| "Voids warranty" / "cloud sync disabled" | Stated harder than the evidence. reMarkable's own wording is that Developer Mode *may affect* warranty and support for problems it causes. Keep our phrasing. |
| "You can turn off dev mode by recovery without permanently harming the device" (proposed FAQ line) | Softer than the truth. Official recovery can erase local data again. Keep our phrasing. |
| "$600 to avoid spending $100" | Category error for the actual audience. They grant sunk hardware in the case *for*, then forget it in the case *against*. The marginal cost is a BYOD license plus the Developer Mode reset. |
| TRMNL X hardware specs (10.3″, dual 6000 mAh, magnet mount) | Unverified from here. Don't repeat in any published copy. |
| Recommendations 1, 2, 5 (one-click installer, documented recovery, "who shouldn't do this") | Already shipped. |

## Their internal inconsistency

They conclude the dominant mindset is *"I want a TRMNL dashboard and incidentally own a
Paper Pro"* — then cite, as evidence, a user saying **"reMarkable was gathering dust.
Nice to give it a new purpose."** That quote supports the opposite branch: a reMarkable
owner repurposing idle hardware. Their best evidence contradicts the verdict attached to
it, so **treat the dominant-use-case question as unresolved.** It's the single most
valuable thing to learn from real users, because it determines whether this is a
reMarkable accessory or a BYOD panel option, and those want different roadmaps.

## Leads worth verifying

- **TRMNL reportedly invites reMarkable BYOD submissions and offers a bounty (~$25) for
  setups.** They read this as evidence the audience is negligible. The likelier read is a
  receptive official channel — which is exactly the "ask TRMNL" item in `PITCH.md` §12.
  Verify on TRMNL's BYOD pages before acting.
- **TRMNL's BYOD page reportedly lists Daylight DC-1 and PineNote**, i.e. TRMNL positions
  itself in the e-ink *screen* ecosystem, not the writing-device ecosystem. If true, that
  shapes how to pitch this to them.
- Their sizing rests on reMarkable's ">1 million sold" figure, which **predates the Paper
  Pro**. Any Paper-Pro-specific number would sharpen the estimate considerably.

## Open questions they couldn't close

- Real measured battery drain for a Paper Pro doing periodic network refreshes. **Our
  battery test can answer this** — that result is publishable and nobody else has it.
- Whether the pairing survives firmware updates in practice, over time.
- Actual community interest, measurable only after release.

## Actions

- [x] Rebut the "sacrificed primary function" objection in the pitch
- [x] Add refresh-interval battery guidance
- [x] Add honest audience sizing and the "don't buy a Paper Pro for this" line
- [ ] Ship Paper-Pro-optimized TRMNL templates for 1620×2160
- [ ] Publish battery-test numbers once collected — this is the differentiating evidence
- [ ] Verify TRMNL's BYOD submission channel and bounty before pitching them
