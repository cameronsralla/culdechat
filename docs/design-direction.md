# Cul-de-Chat: Design Direction
Last Updated: October 4, 2026

Status: **Warm craft locked** (Outfit, grounded stone palette, density A). Refine from real use.

---

## North star
**The modern digital town square for bringing local communities together.**

Private, neighbor-only, warm to use, clear in purpose — help rebuild local community connection.

## Design language: Warm craft

**One line:** Neighborly warmth, craftsman density.

Borrow craft discipline from Linear/Attio/Amie/Resend; keep Cul-de-Chat’s town-square soul. Not dark-native Linear. Not brutalist Resend. Not pastel “toy.”

### Locked
| Topic | Decision |
|---|---|
| Typeface | **Outfit** + JetBrains Mono accents |
| Density | 44px controls / 52px rows on compact; 32px / 40px on `md+` |
| Elevation | Hairline borders + lightness steps (`paper` → `surface` → `raised`). Shadows only on floating chrome |
| Radius | 8 / 12 / 16 (+ pill) |
| Palette | Grounded stone neutrals; deeper desaturated teal as **accent only** — no mint washes filling panels |
| Brand usage | Primary buttons, active nav indicator, focus rings. Not avatar fills, empty-state wells, or whole cards |
| Dark mode | After light is solid |
| Sound | No |
| Emoji in chrome | No |
| Fun | 3 / 5 — press scale; calm when idle |
| Copy | Neighborly and short |

### Surfaces
| Token | Role |
|---|---|
| `paper` | Page background (warm gray stone `#F3F2F0`) |
| `surface` | Panels, lists |
| `raised` | Inputs, popovers, floating chrome |
| `surface-muted` | Wells, selected nav, muted chrome |

### Brand
| Token | Hex | Use |
|---|---|---|
| `brand` | `#008098` | Actions, active accents (matched to logo) |
| `brand-soft` / `brand-wash` | near-neutral tint | Tiny badge fills only — never large panels |

### Control density
| Token | Compact | Desktop (`md+`) |
|---|---|---|
| `control` | 44px | 32px |
| `control-sm` | 36px | 28px |
| `row` | 52px | 40px |

### Motion
- Buttons: `active:scale-[0.98]`
- Sheets/drawers: short spring (when added)
- No decorative idle animation

### What we explicitly do not take
- Linear dark-native monochrome
- Resend black-and-white brutalism
- Cream/mint pastel surfaces that read as toy/kids
- Heavy multi-layer card shadows
- Phone-sized controls on desktop
- Brand-tinted fills for avatars / empty states / whole cards
---

## Reference apps (borrow, don’t copy)
- **Linear** — hairlines, lightness elevation, keyboard density
- **Attio** — compact control baselines, pill badges
- **Amie** — light spring press (dialed to 3/5)
- **Resend** — mono accents for precision data
- **Reflect** — zero-chrome compose later

---

## Feel stack
Structure → Visual system → Interaction → Motion → Delight.
