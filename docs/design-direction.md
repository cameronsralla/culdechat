# Cul-de-Chat: Design Direction
Last Updated: September 30, 2026

Status: **active — best-guess pass applied to the mock.** Working process remains: go item-by-item when refining; vision drives feel ([vision.md](./vision.md)).

---

## Locked so far (Phase 0 — language)

### North star (synthesis)
**The modern digital town square for bringing local communities together.**

A private, neighbor-only social space that makes it easy and pleasant to meet and talk to the people who live around you — exclusive to the community, warm to use, clear in purpose: help rebuild local community connection.

Combo of:
- Cozy private town square you want to open
- Lively neighbor social that feels a bit playful  
…with the deeper why: facilitate neighbor communication and bring people together where they live.

### Fun level
**3 / 5** — warm + light delight on success; not invisible utility, not playful/animated personality-forward.

### Non-negotiables
| Topic | Decision |
|---|---|
| Dark mode | Yes — **after** light mode is solid (tokens later) |
| Sound effects | No |
| Emoji in UI chrome (tabs, badges) | No — more professional/sleek |
| Visual reference | Apple UI + Google clean interfaces, but warmer and more inviting |
| Illustrations / empty-state art | Soft icon wells for now; custom art later if needed |

### Copy voice
**Neighborly and short.** Encourage use in a warm way. Calm when idle — invite, don’t nag.

### Reference apps / philosophy
- **Classic Twitter** — playful enough, strong consistent theme; Cul-de-Chat more polished/clean.
- **Obsidian (philosophy, not look)** — well thought out; gets out of the way when not needed.

---

## Best-guess assumptions (applied Sept 30)

| Question | Assumption | How it shows up |
|---|---|---|
| **A. Exclusivity in UI** | Soft reminders, not loud | “Neighbors only” / “Invite-only” chips on login, shell header, home hero, sidebar |
| **B. Warmth dial** | All lightly | Lagoon teal + warm stone paper, branded wash, HeroBand, softer shadows, Nunito hierarchy |
| **C. Quiet-state tone** | Calm with gentle invite | Empty copy like “Quiet for now…” + soft secondary CTA, not FOMO |
| **D. Dark mode timing** | Later | Light polish only this pass |

### Layout / kit choices this pass
- Home tab labeled **Square**; mobile header title **The square**
- Home opens with a **HeroBand** (atmosphere + exclusivity + purpose)
- People / Messages use **grouped list rows** (less card stack)
- Boards keep cards; joined boards get a brand accent
- You screen teases **About you** (profile richness coming)
- Shell: teal wash, sidebar brand block + “Neighbors only”, signed-in footer

---

## Planned sequence (continued)
1. ~~Finish Phase 0 assumptions~~ (best-guess locked above; revisit if you disagree)
2. Phase 1 — Core loops polish from real use: Messages → Home/post → People/You → Boards → Auth
3. Phase 2 — Continuity (unread, optimistic send, toasts)
4. Phase 3 — Motion system
5. Phase 4 — Light delight
6. Phase 5 — A11y + visual confidence
7. Dark mode tokens when light feels settled

## Feel stack (reminder)
Structure → Visual system → Interaction → Motion → Delight (in that order).
