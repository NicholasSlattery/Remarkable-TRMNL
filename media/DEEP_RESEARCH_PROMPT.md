# Deep Research Prompt — TRMNL × reMarkable as low-distraction devices

**How to use:** paste everything below the horizontal rule into a deep-research tool
(Claude Research, ChatGPT Deep Research, Gemini Deep Research, Perplexity Deep Research).
It is written to stand alone — do not add context, and do not paste project files with it.
The whole point is an independent read that isn't anchored to our own framing.

Two shorter variants are at the bottom for tools with tight input limits.

---

## THE PROMPT

You are a technology analyst with three overlapping specialties: e-paper and
ambient-display hardware, the attention economy and "calm technology" design tradition,
and the practical economics of niche prosumer devices. Produce a rigorous, well-sourced
research report on the question below. Take a position; do not hedge into mush.

### Background you should verify rather than assume

- **TRMNL** is a small e-ink dashboard device and service. It displays a rotating
  "playlist" of screens — calendar, weather, tasks, metrics, RSS, custom plugins — that
  refresh on a schedule rather than on demand. It has a plugin ecosystem, an open
  "BYOD" (bring your own device) program that lets third-party hardware act as a TRMNL
  display, and a "BYOS" (bring your own server) path for self-hosting.
- **reMarkable** makes e-paper tablets for reading and handwriting. The reMarkable Paper
  Pro is its color e-paper model (roughly 1620×2160). The company markets the product
  explicitly on the absence of distraction — no notifications, no app store, no browser
  in normal use. It has a Developer Mode that enables root SSH at the cost of a factory
  reset and reduced device security.
- A community project exists that runs a TRMNL client on a Developer Mode Paper Pro,
  putting TRMNL playlists on the reMarkable's color e-paper display, with scheduled
  refresh, an offline cache, and a reversible return to the stock interface.

Confirm each of these facts independently. Correct anything above that is wrong or has
changed, and say so explicitly — a correction is a finding, not an aside.

### Core research question

**How deep is the actual kinship between TRMNL and reMarkable as low-distraction,
minimalist devices — is it a shared design philosophy or just a shared screen
technology? And given that answer, how useful would running TRMNL software on reMarkable
hardware genuinely be, for whom, and how large is that group?**

Answer in five parts.

### Part 1 — Design philosophy: real kinship or surface resemblance?

Compare the two products as *design positions*, not as spec sheets.

- What does each company actually say about attention, distraction, focus, and calm?
  Quote founders, launch materials, manifestos, and support docs — and note where the
  marketing language outruns the product's real behavior.
- Map both against the established "calm technology" literature — Weiser and Brown's
  original framing, Amber Case's later principles, the ambient-display research
  tradition. Which specific principles does each device satisfy, and which does each
  violate?
- Name the structural differences honestly. Candidates to examine: **push vs. pull**
  (TRMNL pushes a scheduled sequence at you; reMarkable waits for you to open a
  document), **input vs. output** (reMarkable is fundamentally a writing instrument;
  TRMNL has no meaningful input surface), **glanceable vs. immersive** (seconds of
  attention vs. a long focused session), **ambient furniture vs. personal tool**
  (something on a shelf vs. something in your hands).
- Is a dashboard — an object whose entire purpose is to show you information you did not
  ask for at that moment — actually *low-distraction*, or is it a well-mannered
  interruption? Argue both sides and then decide. This is the hinge of the whole report;
  don't skip past it.
- Where does each device sit relative to the broader "digital minimalism" and
  dumbphone/e-ink revival movement (Light Phone, Daylight Computer, Boox, Supernote,
  Kindle Scribe, e-ink desk displays)? Is there a coherent product category forming here,
  and do these two belong to the same one?

### Part 2 — Hardware and technical fit

- Compare the reMarkable Paper Pro's display against TRMNL's own hardware and against
  common BYOD panels: size, resolution, color capability and how color is actually
  achieved, refresh characteristics and ghosting, frontlight, battery, and enclosure.
- Where does the Paper Pro clearly win as a dashboard surface? Where does it clearly lose?
  Consider readability at a distance, viewing angle, whether the device stands up on a
  desk unaided, and what a partial vs. full-panel refresh looks like on this panel.
- **The battery question, treated seriously.** A purpose-built dashboard can idle for
  weeks or months because its firmware does almost nothing between refreshes. A general
  tablet running a full OS with Wi-Fi, a touch stack, and suspend/resume cycles cannot.
  Find real figures for both. How much does scheduled RTC wake-and-refresh actually
  recover, and what refresh interval makes a tablet-based dashboard practical — hourly?
  A few times a day? Does the answer change if it stays plugged in, and does keeping an
  e-paper tablet permanently plugged in create its own problems?
- What does it cost to *keep* a modified device working — firmware updates that break
  private interfaces, security posture on a root-SSH device, the fact that a factory
  reset is required to enter Developer Mode and official recovery is required to leave it?

### Part 3 — Usefulness: the honest case for and against

Build the strongest version of each side, then judge.

**The case for.** Argue it properly: dramatically better screen than typical BYOD panels;
color; hardware people already own sitting idle most of the day; one fewer object on the
desk; no new purchase; a device already trusted to be quiet.

**The case against.** Argue it just as hard: dedicated dashboards cost far less than a
reMarkable and are designed for exactly this; a reMarkable in dashboard mode isn't
available for its primary job; Developer Mode's factory reset and security reduction are
paid before any benefit arrives; firmware updates are a standing threat; a tablet's
battery life in dashboard use may be measured in days rather than months.

Then answer directly: **is the dominant real-world use "my reMarkable is now also a
dashboard when idle," or "I want a TRMNL dashboard and this is the panel I happen to
have"?** These imply very different products, different feature priorities, and different
audiences. Evidence for which one dominates is the most valuable thing this report can
produce.

### Part 4 — Audience and market size

- Estimate, with stated methodology and explicit uncertainty, the size of the overlap:
  people who own a reMarkable Paper Pro, are willing to enable Developer Mode, and want a
  dashboard. Show your arithmetic. A defensible range with visible assumptions beats a
  confident single number.
- Characterize that person concretely — occupation, existing tool stack, what else they've
  modified, what they'd actually put on the dashboard.
- Find evidence of latent demand: reMarkable community threads asking for dashboard,
  calendar, or always-on display features; TRMNL community discussion of reMarkable as a
  BYOD target; existing reMarkable dashboard or "e-ink status board" projects and how much
  traction they got. Search Reddit (r/RemarkableTablet, r/eink, r/selfhosted), the
  reMarkable community forum, Hacker News, GitHub, and TRMNL's own community spaces.
- Which adjacent audiences are *larger* and might matter more — self-hosters, home-lab
  operators, quantified-self users, ADHD and focus-tool communities, calm-tech
  enthusiasts?
- Are there competing or prior projects covering this ground? If so, what happened to
  them, and what does their trajectory predict?

### Part 5 — Verdict and recommendations

- Give a clear verdict on the philosophical kinship: **deep, partial, or superficial?**
  Defend the choice in a paragraph.
- Give a clear verdict on usefulness: is this a genuinely valuable pairing, a clever niche
  toy, or a mismatch that looks good in a photograph? One sentence, then the defense.
- Rank the three or four use cases where the pairing is *most* defensible, and the two or
  three where someone should be told to buy a dedicated dashboard instead.
- Identify the single biggest technical or product risk to this pairing being useful over
  a two-year horizon.
- List the five feature or documentation priorities that would most increase real-world
  usefulness, ranked by impact per unit of effort.
- Name three framing or positioning mistakes a project like this is likely to make when
  describing itself, and what to say instead.

### Sources — priority order

1. Primary product documentation, API docs, spec sheets, and support articles from both
   companies
2. Founder and company statements — interviews, launch posts, changelogs, podcast
   appearances
3. Community discussion with real usage detail: Reddit, official forums, Hacker News,
   Discord and Slack archives that are publicly indexed, GitHub issues
4. Independent hands-on reviews with measured figures, especially battery and refresh
5. Academic and design literature on calm technology, ambient displays, and attention
6. Comparable products and projects — other BYOD panels, e-ink dashboards, other
   reMarkable modifications

### Ground rules

- **Cite everything.** Every non-obvious factual claim gets a source link.
- **Never invent a number.** No fabricated sales figures, user counts, or battery
  measurements. If a figure isn't publicly available, write "not publicly available" and
  give your reasoning for any estimate you construct from it.
- **Separate evidence from inference.** Mark estimates and opinions as such.
- **Flag staleness.** Note the date of anything that may have changed, especially firmware
  behavior, pricing, and BYOD program terms.
- **Contradict the premise if the evidence says to.** If the kinship is mostly superficial,
  or if this pairing is a bad idea, say so plainly and support it. A well-argued negative
  finding is the most useful possible outcome.
- **Don't pad.** Every section should contain something a well-informed reader didn't
  already know.

### Output format

1. **Verdict** — under 200 words, both verdicts up front, no throat-clearing
2. **Philosophy comparison** — including a table mapping each device against calm-tech
   principles
3. **Technical fit** — including a hardware comparison table with sourced figures
4. **Usefulness** — steelmanned case for, steelmanned case against, judgment
5. **Audience and market** — with visible arithmetic and stated uncertainty
6. **Recommendations** — ranked, specific, actionable
7. **Open questions** — what you could not resolve, and what evidence would resolve it
8. **Sources** — annotated, grouped by the priority tiers above

Target 2,500–4,000 words plus tables. Prefer density over length.

---

## Short variant (for tools with input limits)

> Research how similar TRMNL (the e-ink dashboard device and BYOD/BYOS service) and
> reMarkable (the e-paper writing tablet, especially the color Paper Pro) really are as
> minimalist, low-distraction devices — shared design philosophy, or just shared screen
> technology? Compare what each company says about attention and focus against how the
> products actually behave, and map both against the calm-technology literature. Address
> the structural differences directly: push vs. pull, input vs. output, glanceable vs.
> immersive. Then assess how useful running TRMNL software on a reMarkable Paper Pro
> would actually be — the case for (far better screen than typical BYOD panels, color,
> hardware people already own and that sits idle most of the day) and the case against
> (dedicated dashboards cost much less, Developer Mode requires a factory reset and
> reduces security, firmware updates can break it, tablet battery life in dashboard use
> is likely days not months). Estimate the size of the overlap audience with visible
> arithmetic, find evidence of latent demand in reMarkable and TRMNL community
> discussion, and end with a clear verdict on both kinship and usefulness plus five
> ranked priorities that would make the pairing more useful. Cite every claim, invent no
> numbers, and contradict the premise if the evidence supports doing so.

## One-paragraph variant

> Are TRMNL and reMarkable genuinely kindred low-distraction devices or just two products
> that happen to use e-paper? Compare their design philosophies against the calm-tech
> literature and against how each actually behaves, then judge how useful TRMNL software
> running on a reMarkable Paper Pro would be — who wants it, how many of them exist, what
> Developer Mode and battery life really cost, and whether a dedicated dashboard is simply
> the better buy. Cite sources, invent no figures, and give a clear verdict.
