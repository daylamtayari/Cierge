# Cierge Design System

This document defines the visual language, design tokens, component patterns, and UX conventions for the Cierge frontend. Follow it exactly when generating any UI code. The reference implementation is `server/web/mockup.html`.

---

## Brand

**Cierge** (short for concierge) is a restaurant reservation scheduling platform. The interface should feel like a trusted concierge at a fine hotel — assured, refined, effortless. It communicates "this is handled" rather than "look at all this complexity."

**Tagline**: "Your reservation, handled."

---

## Design Principles

1. **Clarity over cleverness** — Every element must be immediately understood by non-technical users. Use labels, not icons alone. Obvious over subtle.
2. **Calm confidence** — No visual urgency, no unnecessary alerts. Status is communicated clearly but quietly.
3. **Warmth, not sterility** — Tinted neutrals, the signature red, and generous spacing create a space that feels human, not clinical.
4. **Progressive disclosure** — Show what matters now, reveal details on demand.
5. **Works everywhere** — Mobile is a first-class experience. Core flows must be effortless on a phone.

---

## What This Must NOT Look Like

- **Resy** (dark/moody/editorial)
- **OpenTable** (bright/commercial/busy)
- **Generic SaaS dashboards** (cards-in-cards, gradient metrics, icon grids)
- **AI-generated UI** — no cyan-on-dark, no purple gradients, no glassmorphism, no glassmorphic cards with blur/glow, no rounded rectangles with generic drop shadows, no gradient text, no card grids with icons above headings, no hero metric layouts. If someone's first reaction is "an AI made this," the design has failed.

---

## Typography

### Font Stack

| Role | Font | Fallback | Weight |
|------|------|----------|--------|
| Display (headings, restaurant names) | DM Serif Text | Georgia, Times New Roman, serif | 400 (regular) |
| Body (UI text, labels, paragraphs) | Plus Jakarta Sans | system-ui, -apple-system, sans-serif | 400, 500, 600 |
| Code (API keys, logs, confirmations) | SF Mono | Cascadia Code, Consolas, monospace | 400 |

### Loading

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=DM+Serif+Text:ital@0;1&family=Plus+Jakarta+Sans:wght@400;500;600&display=swap" rel="stylesheet">
```

### Type Scale (1.25 ratio, fluid)

| Token | Size | Usage |
|-------|------|-------|
| `--text-xs` | clamp(0.7rem, 0.65rem + 0.2vw, 0.75rem) | Captions, metadata, timestamps |
| `--text-sm` | clamp(0.8rem, 0.75rem + 0.2vw, 0.875rem) | Secondary text, labels, buttons |
| `--text-base` | clamp(0.9rem, 0.85rem + 0.2vw, 1rem) | Body text, form inputs |
| `--text-lg` | clamp(1.1rem, 1rem + 0.3vw, 1.25rem) | Section emphasis, restaurant names in lists, topbar brand |
| `--text-xl` | clamp(1.4rem, 1.2rem + 0.5vw, 1.75rem) | Detail page headings |
| `--text-2xl` | clamp(1.8rem, 1.5rem + 0.8vw, 2.25rem) | Page titles, login brand |

### Heading Styles

- **Page heading** (`heading-page`): `font-family: var(--font-display)`, `font-size: var(--text-2xl)`, `color: var(--ink)`, `letter-spacing: -0.02em`, `line-height: 1.2`
- **Section heading** (`heading-section`): `font-family: var(--font-body)`, `font-size: var(--text-base)`, `font-weight: 600`, `color: var(--ink)`
- **Restaurant names** in job rows: `font-family: var(--font-display)`, `font-size: var(--text-lg)`, `color: var(--ink)`

### Key Rule

The display serif (DM Serif Text) is reserved for: page titles, restaurant names, confirmation numbers, and the brand mark. Everything else uses Plus Jakarta Sans. Never use monospace for UI elements — it is only for literal code/keys/logs.

---

## Color System

All colors use **OKLCH** for perceptual uniformity. All neutrals are **tinted toward hue 18** (warm red) for subconscious cohesion.

### Light Mode

```css
/* Signature red */
--red:          oklch(52% 0.14 18);   /* Primary actions, brand, active states */
--red-light:    oklch(62% 0.12 18);   /* Hover state for primary */
--red-subtle:   oklch(94% 0.02 18);   /* Selected backgrounds, initials bg */
--red-dark:     oklch(38% 0.10 18);   /* Active/pressed state */

/* Warm neutrals */
--ink:            oklch(20% 0.01 18);   /* Headings, strongest text */
--text:           oklch(28% 0.01 18);   /* Body text */
--text-secondary: oklch(45% 0.008 18);  /* Supporting text */
--text-tertiary:  oklch(62% 0.006 18);  /* Metadata, placeholders, labels in detail grids */
--border:         oklch(88% 0.006 18);  /* Primary borders */
--border-subtle:  oklch(92% 0.004 18);  /* Subtle dividers, job row borders */
--surface:        oklch(99% 0.002 18);  /* Topbar, main surface */
--surface-raised: oklch(100% 0 0);      /* Cards, job rows, inputs */
--bg:             oklch(96.5% 0.004 18); /* Page background */

/* Semantic */
--confirmed:    oklch(52% 0.12 155);   /* Success/confirmed state */
--confirmed-bg: oklch(95% 0.03 155);   /* Success background */
--failed:       oklch(52% 0.14 25);    /* Error/failed state */
--failed-bg:    oklch(95% 0.03 25);    /* Error background */
--pending:      oklch(52% 0.10 250);   /* Scheduled/pending state */
--pending-bg:   oklch(95% 0.03 250);   /* Pending background */
--muted-bg:     oklch(94% 0.004 18);   /* Cancelled, disabled backgrounds */
```

### Dark Mode

Applied via `@media (prefers-color-scheme: dark)`. Key shifts:

```css
--red:          oklch(65% 0.13 18);   /* Brighter for contrast */
--red-light:    oklch(72% 0.11 18);
--red-subtle:   oklch(25% 0.04 18);   /* Dark tinted background */
--red-dark:     oklch(55% 0.10 18);

--ink:            oklch(95% 0.005 18);
--text:           oklch(90% 0.005 18);
--text-secondary: oklch(68% 0.006 18);
--text-tertiary:  oklch(52% 0.005 18);
--border:         oklch(30% 0.008 18);
--border-subtle:  oklch(25% 0.006 18);
--surface:        oklch(18% 0.006 18);
--surface-raised: oklch(22% 0.008 18);
--bg:             oklch(14% 0.005 18);

--confirmed:    oklch(68% 0.12 155);
--confirmed-bg: oklch(22% 0.04 155);
--failed:       oklch(68% 0.14 25);
--failed-bg:    oklch(22% 0.04 25);
--pending:      oklch(68% 0.10 250);
--pending-bg:   oklch(22% 0.04 250);
--muted-bg:     oklch(22% 0.005 18);
```

### Usage Rules

- **Red appears in**: brand text, primary buttons, active tab underline, selected option borders, selected restaurant border, job row hover accent, "add new" links. Nowhere else.
- **Never use red for errors.** Errors use `--failed` (hue 25, which is orange-red, distinct from the brand red at hue 18).
- **Never use pure black or pure white.** The darkest color is `--ink` at 20% lightness; the lightest is `--surface-raised` at 100% but only for raised elements.
- **All neutral colors carry the warm tint.** When adding new neutrals, always include `0.005-0.01` chroma at hue 18.

---

## Spacing

4px base unit. Use these tokens exclusively:

| Token | Value | Usage |
|-------|-------|-------|
| `--sp-1` | 4px | Tight gaps (between tag dot and text) |
| `--sp-2` | 8px | Small gaps (button icon gap, field-row gaps) |
| `--sp-3` | 12px | Medium gaps (padding inside options, nav gaps) |
| `--sp-4` | 16px | Standard padding (inputs, job rows, fields between) |
| `--sp-6` | 24px | Section internal spacing, container side padding |
| `--sp-8` | 32px | Section gaps, page header margin-bottom, container top padding |
| `--sp-12` | 48px | Large separations (login brand to form, between tab sets) |
| `--sp-16` | 64px | Page-level breathing room |

### Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `--radius` | 6px | Inputs, buttons, small elements |
| `--radius-lg` | 10px | Job rows, banners, selected restaurant |

---

## Layout

### Container

Single centered container: `max-width: 720px`, `margin: 0 auto`, padding `var(--sp-8) var(--sp-6)`.

No narrow/wide variants. One container width for all pages.

### Topbar

- Height: 56px, sticky
- Background: `var(--surface)` with bottom border
- Brand text: `var(--font-display)` in `var(--red)`
- Nav links: `var(--font-body)`, `var(--text-sm)`, weight 500
- Active link: `aria-current="page"` attribute, styled with `var(--bg)` background
- Right side: initials circle (30px, `var(--red-subtle)` bg, `var(--red)` text)

### Page Structure

```
<div class="shell">
  <header class="topbar">...</header>
  <main class="container">
    <h1 class="heading-page">...</h1>
    <section class="section">...</section>
  </main>
</div>
```

---

## Components

### Buttons

| Variant | Background | Text | Border | Usage |
|---------|-----------|------|--------|-------|
| `btn-primary` | `var(--red)` | white | `var(--red)` | Main CTA per page (one only) |
| `btn-secondary` | `var(--surface-raised)` | `var(--text)` | `var(--border)` | Secondary actions |
| `btn-subtle` | transparent | `var(--text-secondary)` | none | Tertiary actions (change, back) |
| `btn-danger-outline` | transparent | `var(--failed)` | `var(--border)` | Destructive actions (cancel booking) |

- All buttons: `font-size: var(--text-sm)`, `font-weight: 500`, `border-radius: var(--radius)`, `padding: 10px var(--sp-4)`
- Small variant (`btn-sm`): `padding: 6px var(--sp-3)`, `font-size: var(--text-xs)`
- Touch devices (`@media (pointer: coarse)`): increase padding to 12px / 10px
- Focus: `outline: 2px solid var(--red); outline-offset: 2px`
- Transitions: 120ms with `var(--ease-out)`

**Label rules**: Use specific action verbs. "Schedule this booking" not "Submit". "Update password" not "Save". "Cancel this booking" not "Cancel".

### Status Tags

Pill-shaped with a colored dot (6px circle via `::before` pseudo-element).

| Status | Class | Background | Text Color |
|--------|-------|-----------|------------|
| Scheduled | `tag-scheduled` | `var(--pending-bg)` | `var(--pending)` |
| Confirmed | `tag-confirmed` | `var(--confirmed-bg)` | `var(--confirmed)` |
| Failed | `tag-failed` | `var(--failed-bg)` | `var(--failed)` |
| Cancelled | `tag-cancelled` | `var(--muted-bg)` | `var(--text-tertiary)` |

Style: `padding: 3px 10px`, `border-radius: 100px`, `font-size: var(--text-xs)`, `font-weight: 500`.

### Job Rows

Each booking in the list is an `<article class="job-row">`:
- Background: `var(--surface-raised)`, border: `var(--border-subtle)`, radius: `var(--radius-lg)`
- Padding: `var(--sp-4)` with extra left padding `var(--sp-6)` for the accent line
- **Left accent line**: 3px wide, `var(--border)` default, turns `var(--red)` on hover. Positioned via `::before` pseudo-element
- Hover: border darkens, row lifts 1px (`translateY(-1px)`) on devices with hover capability
- Restaurant name: display font. Details: secondary text.
- Right side: status tag + date stacked

### Forms

- Labels (`field-label`): `var(--text-sm)`, weight 500, margin-bottom `var(--sp-2)`
- Inputs (`field-input`): padding `10px var(--sp-3)`, border `var(--border)`, radius `var(--radius)`, `font-size: var(--text-base)`
- Focus: `border-color: var(--red)` + `outline: 2px solid var(--red)`
- Hints (`field-hint`): `var(--text-xs)`, `var(--text-tertiary)`, margin-top `var(--sp-1)`
- Row layout (`field-row`): flex with `gap: var(--sp-3)`, stacks on mobile

### Tabs

- Container (`tab-bar`): flex with `gap: var(--sp-6)`, bottom border
- Buttons (`tab-btn`): no background, `var(--text-tertiary)` default, `var(--text-sm)` weight 500
- Active: `aria-selected="true"`, color `var(--red)`, 2px bottom border in `var(--red)`

### Detail Grids

Use semantic `<dl>` with CSS grid:
```css
.details {
  display: grid;
  grid-template-columns: 140px 1fr;
  gap: var(--sp-2) var(--sp-4);
  font-size: var(--text-sm);
}
.details dt { color: var(--text-tertiary); }
.details dd { color: var(--text); }
```

On mobile: columns shrink to `120px 1fr`.

### Banners

Full-width blocks for job results:
- **Confirmed**: `var(--confirmed-bg)` background, 1px border. Label in `var(--confirmed)`, confirmation number in display font at `var(--text-xl)`.
- **Failed**: `var(--failed-bg)` background, 1px border. Label + error message in `var(--failed)`.
- Padding: `var(--sp-6)`, radius: `var(--radius-lg)`.

### Summary Block

Used in the booking confirmation step:
- Left border: 3px solid `var(--red)` (not a full card border)
- Asymmetric radius: `2px var(--radius-lg) var(--radius-lg) 2px`
- Contains detail grid + primary CTA

### Log Viewer

- Background: `oklch(15% 0.005 18)` (near-black, warm-tinted)
- Text: `oklch(80% 0 0)`
- Font: monospace, `var(--text-xs)`, line-height 1.7
- Max height: 220px with overflow scroll
- Line colors: `.info` = `oklch(70% 0.06 250)`, `.ok` = `oklch(70% 0.1 155)`, `.err` = `oklch(70% 0.12 25)`

### Option Selector (Drop Configs)

Stacked selectable items:
- Border: `var(--border)`, radius: `var(--radius)`, padding: `var(--sp-3) var(--sp-4)`
- Selected (`aria-selected="true"`): border `var(--red)`, background `var(--red-subtle)`
- Contains label (weight 500) and meta text (tertiary)

---

## Terminology

Use these terms consistently. Never mix.

| Internal/Technical | User-Facing |
|-------------------|-------------|
| Job | Booking |
| Success (status) | Confirmed |
| Drop config | "When do reservations open?" / release schedule |
| Confidence (number) | "Used X times" |
| Party size value | "X guests" |
| Execute / run | "Will run at" / "Ran at" |
| Authentication token | Sign in / Sign out |

---

## Accessibility

- **Focus rings**: 2px solid `var(--red)`, 2px offset. Never remove focus outlines without `focus-visible` replacement.
- **Reduced motion**: Wrap all animations in `@media (prefers-reduced-motion: reduce)` with near-zero durations.
- **Touch targets**: `@media (pointer: coarse)` increases button padding. Minimum 44px tap targets.
- **Safe areas**: Respect `env(safe-area-inset-*)` for notched devices.
- **Semantic HTML**: Use `<article>` for job rows, `<section>` for content groups, `<dl>` for key-value pairs, `<nav>` for navigation, `aria-current="page"` for active nav, `aria-selected` for tabs and options.
- **Labels**: Every form input has an associated `<label>` with `for` attribute. Never use placeholder as label.
- **Color contrast**: All text meets WCAG AA (4.5:1 body, 3:1 large text). Test both light and dark mode.

---

## Motion

- **Easing**: `cubic-bezier(0.25, 1, 0.5, 1)` (`--ease-out`) for all transitions
- **Duration**: 120ms for interactive feedback (hover, focus), 150ms for layout hints (job row lift), 300ms for page entrance
- **Page entrance**: `fade-up` animation (opacity 0 → 1, translateY 8px → 0) on `.container` when page becomes active
- **Only animate**: `transform` and `opacity`. Never animate layout properties.
- **Job row hover**: `border-color` transition + `translateY(-1px)` + accent line color change. Only on `@media (hover: hover)`.

---

## Responsive

### Breakpoint

Single breakpoint at **640px**. Mobile-first: base styles are mobile, `min-width: 640px` adds desktop refinements.

### Mobile Adaptations

- Container padding reduces to `var(--sp-6) var(--sp-4)`
- `field-row` stacks vertically
- Job rows stack (info on top, status/date below in a horizontal row)
- Detail grid columns: `120px 1fr`
- Detail headers stack vertically

### Input Method

- `@media (pointer: coarse)`: Larger touch targets on buttons and inputs
- `@media (hover: hover)`: Hover effects (job row lift) only for pointer devices
- `@media (hover: none)`: No hover states; rely on active states

---

## Page Inventory

| Page | URL Pattern | Purpose |
|------|------------|---------|
| Login | `/login` | Email/password sign-in |
| Dashboard | `/` | List of upcoming and past bookings with tab toggle |
| New Booking | `/bookings/new` | 3-step flow: restaurant → details → confirm |
| Booking Detail | `/bookings/:id` | Status, details, results. Varies by state (scheduled/confirmed/failed/cancelled) |
| Settings | `/settings` | Platform connections, API key, password, account |

---

## File References

- **Mockup**: `server/web/mockup.html` — Complete interactive reference with all pages
- **Design context**: `.impeccable.md` — Brand personality, users, principles
- **This file**: `server/web/DESIGN_SYSTEM.md` — Token and component specification
