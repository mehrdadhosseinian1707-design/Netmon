# NOC System Design - Comprehensive Style Guide

## Overview
This style guide documents the complete design system for the NOC (Network Operations Center) Enterprise Network Management interface by IndoBSD. The design features a futuristic dark theme with glassmorphism effects, gradient accents, and a sophisticated color palette.

---

## Typography

### Font Families

**Primary Font: Inter**
- Usage: Body text, UI elements, labels, values
- Fallback: `sans-serif`
- Declaration: `font-family: 'Inter', sans-serif`

**Secondary Font: Space Grotesk**
- Usage: Headings, titles, emphasis elements
- Fallback: `sans-serif`
- Declaration: `font-family: 'Space Grotesk', sans-serif`

**Monospace Font**
- Usage: Code, technical data, IDs
- Stack: `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace`

### Font Sizes

| Element Type | Size | rem | Line Height |
|-------------|------|-----|-------------|
| Extra Small | 8px | 0.5rem | - |
| Small | 9px | 0.5625rem | - |
| XS Text | 10px | 0.625rem | - |
| Fine Print | 11px | 0.6875rem | - |
| Small Text | 12px | 0.75rem | 1.25 |
| Body Text | 14px | 0.875rem | 1.5 |
| Base | 16px | 1rem | 1.5 |
| Large | 18px | 1.125rem | 1.75 |
| XL | 20px | 1.25rem | 1.75 |
| 2XL | 24px | 1.5rem | 2rem |
| 3XL | 30px | 1.875rem | 2.25rem |
| 4XL | 36px | 2.25rem | 2.5rem |
| 5XL | 48px | 3rem | 1 |
| 6XL | 60px | 3.75rem | 1 |
| Hero | 54.4px | 3.4rem | 1 |

### Font Weights

- Regular: 400
- Medium: 500
- Semibold: 600
- Bold: 700
- Extrabold: 800

---

## Color Palette

### Base Colors

**Background Colors**
- Primary Background: `#020617` (rgb(2, 6, 23))
- Secondary Background: `#030b1a` (rgb(3, 11, 26))
- Tertiary Background: `#0c1628` (rgb(12, 22, 40))

**Text Colors**
- Primary Text: `#fff` (white)
- Secondary Text: `#e5e7eb` (rgb(229, 231, 235))
- Muted Text: `#9ca3af` (rgb(156, 163, 175))
- Tertiary Text: `#64748b` (rgb(100, 116, 139))

### Accent Colors

**Cyan/Aqua (Primary Accent)**
- Base: `#22d3ee` (rgb(34, 211, 238))
- Light: `#67e8f9` (rgb(103, 232, 249))
- Dark: `#06b6d4` (rgb(6, 182, 212))
- Variants:
  - 80%: `rgba(34, 211, 238, 0.8)`
  - 60%: `rgba(34, 211, 238, 0.6)`
  - 50%: `rgba(34, 211, 238, 0.5)`
  - 30%: `rgba(34, 211, 238, 0.3)`
  - 28%: `rgba(34, 211, 238, 0.28)`
  - 25%: `rgba(34, 211, 238, 0.25)`
  - 20%: `rgba(34, 211, 238, 0.2)`
  - 15%: `rgba(34, 211, 238, 0.15)`
  - 12%: `rgba(34, 211, 238, 0.12)`
  - 10%: `rgba(34, 211, 238, 0.1)`
  - 6%: `rgba(34, 211, 238, 0.06)`
  - 5%: `rgba(34, 211, 238, 0.05)`
  - 4%: `rgba(34, 211, 238, 0.04)`
  - 3%: `rgba(34, 211, 238, 0.03)`

**Violet/Purple**
- Violet 400: `#a78bfa` (rgb(167, 139, 250))
- Violet 500: `#8b5cf6` (rgb(139, 92, 246))
- Purple 400: `#c084fc` (rgb(192, 132, 252))
- Purple 300: `#d8b4fe` (rgb(216, 180, 254))
- Fuchsia: `#e879f9` (rgb(232, 121, 249))
- Variants:
  - 90%: `rgba(167, 139, 250, 0.9)`
  - 80%: `rgba(167, 139, 250, 0.8)`
  - 60%: `rgba(167, 139, 250, 0.6)`
  - 55%: `rgba(167, 139, 250, 0.55)`
  - 45%: `rgba(167, 139, 250, 0.45)`
  - 35%: `rgba(167, 139, 250, 0.35)`
  - 28%: `rgba(167, 139, 250, 0.28)`
  - 25%: `rgba(167, 139, 250, 0.25)`
  - 20%: `rgba(167, 139, 250, 0.2)`
  - 18%: `rgba(167, 139, 250, 0.18)`
  - 12%: `rgba(167, 139, 250, 0.12)`
  - 10%: `rgba(167, 139, 250, 0.1)`
  - 8%: `rgba(167, 139, 250, 0.08)`
  - 7%: `rgba(167, 139, 250, 0.07)`
  - 5%: `rgba(167, 139, 250, 0.05)`

**Blue**
- Blue 400: `#60a5fa` (rgb(96, 165, 250))
- Blue 500: `#3b82f6` (rgb(59, 130, 246))
- Indigo 400: `#818cf8` (rgb(129, 140, 248))
- Sky 400: `#38bdf8` (rgb(56, 189, 248))
- Variants (Blue 400):
  - 80%: `rgba(96, 165, 250, 0.8)`
  - 70%: `rgba(96, 165, 250, 0.7)`
  - 60%: `rgba(96, 165, 250, 0.6)`
  - 55%: `rgba(96, 165, 250, 0.55)`
  - 45%: `rgba(96, 165, 250, 0.45)`
  - 40%: `rgba(96, 165, 250, 0.4)`
  - 30%: `rgba(96, 165, 250, 0.3)`
  - 25%: `rgba(96, 165, 250, 0.25)`
  - 20%: `rgba(96, 165, 250, 0.2)`
  - 15%: `rgba(96, 165, 250, 0.15)`
  - 14%: `rgba(96, 165, 250, 0.14)`
  - 10%: `rgba(96, 165, 250, 0.1)`
  - 5%: `rgba(96, 165, 250, 0.05)`
  - 4%: `rgba(96, 165, 250, 0.04)`

**Green (Success)**
- Green 400: `#4ade80` (rgb(74, 222, 128))
- Green 500: `#22c55e` (rgb(34, 197, 94))
- Teal 400: `#2dd4bf` (rgb(45, 212, 191))
- Variants (Green 400):
  - 90%: `rgba(74, 222, 128, 0.9)`
  - 80%: `rgba(74, 222, 128, 0.8)`
  - 70%: `rgba(74, 222, 128, 0.7)`
  - 60%: `rgba(74, 222, 128, 0.6)`
  - 55%: `rgba(74, 222, 128, 0.55)`
  - 50%: `rgba(74, 222, 128, 0.5)`
  - 40%: `rgba(74, 222, 128, 0.4)`
  - 35%: `rgba(74, 222, 128, 0.35)`
  - 32%: `rgba(74, 222, 128, 0.32)`
  - 30%: `rgba(74, 222, 128, 0.3)`
  - 25%: `rgba(74, 222, 128, 0.25)`
  - 20%: `rgba(74, 222, 128, 0.2)`
  - 15%: `rgba(74, 222, 128, 0.15)`
  - 10%: `rgba(74, 222, 128, 0.1)`
  - 6%: `rgba(74, 222, 128, 0.06)`
  - 5%: `rgba(74, 222, 128, 0.05)`

**Yellow/Amber (Warning)**
- Yellow 400: `#facc15` (rgb(250, 204, 21))
- Yellow 300: `#fcd34d` (rgb(252, 211, 77))
- Amber 500: `#f59e0b` (rgb(245, 158, 11))
- Amber 300: `#fbbf24` (rgb(251, 191, 36))
- Variants (Yellow 400):
  - 90%: `rgba(250, 204, 21, 0.9)`
  - 80%: `rgba(250, 204, 21, 0.8)`
  - 60%: `rgba(250, 204, 21, 0.6)`
  - 35%: `rgba(250, 204, 21, 0.35)`
  - 30%: `rgba(250, 204, 21, 0.3)`
  - 25%: `rgba(250, 204, 21, 0.25)`
  - 20%: `rgba(250, 204, 21, 0.2)`
  - 15%: `rgba(250, 204, 21, 0.15)`
  - 10%: `rgba(250, 204, 21, 0.1)`
  - 5%: `rgba(250, 204, 21, 0.05)`

**Orange**
- Orange 400: `#fb923c` (rgb(251, 146, 60))
- Orange 500: `#f97316` (rgb(249, 115, 22))
- Orange 300: `#fdba74` (rgb(253, 186, 116))
- Variants:
  - 90%: `rgba(251, 146, 60, 0.9)`
  - 30%: `rgba(251, 146, 60, 0.3)`
  - 20%: `rgba(251, 146, 60, 0.2)`
  - 10%: `rgba(251, 146, 60, 0.1)`
  - 5%: `rgba(251, 146, 60, 0.05)`

**Red (Error/Critical)**
- Red 400: `#f87171` (rgb(248, 113, 113))
- Red 500: `#ef4444` (rgb(239, 68, 68))
- Rose 400: `#fb7185` (rgb(251, 113, 133))
- Variants (Red 400):
  - 90%: `rgba(248, 113, 113, 0.9)`
  - 80%: `rgba(248, 113, 113, 0.8)`
  - 70%: `rgba(248, 113, 113, 0.7)`
  - 60%: `rgba(248, 113, 113, 0.6)`
  - 55%: `rgba(248, 113, 113, 0.55)`
  - 35%: `rgba(248, 113, 113, 0.35)`
  - 30%: `rgba(248, 113, 113, 0.3)`
  - 25%: `rgba(248, 113, 113, 0.25)`
  - 20%: `rgba(248, 113, 113, 0.2)`
  - 15%: `rgba(248, 113, 113, 0.15)`
  - 10%: `rgba(248, 113, 113, 0.1)`
  - 5%: `rgba(248, 113, 113, 0.05)`
- Variants (Red 500):
  - 80%: `rgba(239, 68, 68, 0.8)`
  - 60%: `rgba(239, 68, 68, 0.6)`
  - 50%: `rgba(239, 68, 68, 0.5)`
  - 30%: `rgba(239, 68, 68, 0.3)`
  - 20%: `rgba(239, 68, 68, 0.2)`
  - 10%: `rgba(239, 68, 68, 0.1)`
  - 9%: `rgba(239, 68, 68, 0.09)`
  - 7%: `rgba(239, 68, 68, 0.07)`

**White (Overlay/Glass Effect)**
- Variants:
  - 70%: `rgba(255, 255, 255, 0.7)`
  - 20%: `rgba(255, 255, 255, 0.2)`
  - 10%: `rgba(255, 255, 255, 0.1)`
  - 7%: `rgba(255, 255, 255, 0.07)`
  - 6%: `rgba(255, 255, 255, 0.06)`
  - 5%: `rgba(255, 255, 255, 0.05)`
  - 4%: `rgba(255, 255, 255, 0.04)`
  - 3%: `rgba(255, 255, 255, 0.03)`
  - 2%: `rgba(255, 255, 255, 0.02)`
  - 1.8%: `rgba(255, 255, 255, 0.018)`

**Black (Shadow/Overlay)**
- Variants:
  - 70%: `rgba(0, 0, 0, 0.7)`
  - 60%: `rgba(0, 0, 0, 0.6)`
  - 50%: `rgba(0, 0, 0, 0.5)`
  - 40%: `rgba(0, 0, 0, 0.4)`
  - 30%: `rgba(0, 0, 0, 0.3)`
  - 8%: `rgba(0, 0, 0, 0.08)`

---

## Spacing System

### Padding Scale

| Class | Value | rem | px |
|-------|-------|-----|----|
| p-0.5 | 0.125rem | 0.125 | 2px |
| p-1 | 0.25rem | 0.25 | 4px |
| p-2 | 0.5rem | 0.5 | 8px |
| p-2.5 | 0.625rem | 0.625 | 10px |
| p-3 | 0.75rem | 0.75 | 12px |
| p-4 | 1rem | 1 | 16px |
| p-5 | 1.25rem | 1.25 | 20px |
| p-6 | 1.5rem | 1.5 | 24px |
| p-8 | 2rem | 2 | 32px |
| p-10 | 2.5rem | 2.5 | 40px |
| p-12 | 3rem | 3 | 48px |
| p-16 | 4rem | 4 | 64px |

### Common Padding Patterns

**Horizontal Padding**
- Small: `px-4` (16px)
- Medium: `px-6` (24px)
- Large: `px-8` (32px)
- XLarge: `px-16` (64px)
- Container: `0 48px`

**Vertical Padding**
- Small: `py-2` (8px)
- Medium: `py-4` (16px)
- Large: `py-6` (24px)
- XLarge: `py-8` (32px)

### Margin Scale

Follows the same scale as padding (0.125rem to 4rem)

---

## Border Styles

### Border Radius

| Usage | Value | rem | px |
|-------|-------|-----|----|
| None | 0 | 0 | 0px |
| Small | 0.125rem | 0.125 | 2px |
| Default | 0.25rem | 0.25 | 4px |
| Medium | 0.375rem | 0.375 | 6px |
| Large | 0.5rem | 0.5 | 8px |
| XL | 0.75rem | 0.75 | 12px |
| 2XL | 1rem | 1 | 16px |
| Full | 9999px | - | Full circle |

### Common Border Radius Usage

- **Cards**: `0.75rem` (12px) or `1rem` (16px)
- **Buttons**: `0.375rem` (6px) or `0.5rem` (8px)
- **Inputs**: `0.5rem` (8px)
- **Badges**: `0.25rem` (4px) or `9999px` (full)
- **Avatars**: `9999px` (full circle)

### Border Colors

**Primary Borders**
- Cyan variants: `rgba(34, 211, 238, 0.3)` to `rgba(34, 211, 238, 0.1)`
- Default: `rgba(255, 255, 255, 0.1)`

**Status Borders**
- Success: `rgba(74, 222, 128, 0.3)` to `rgba(74, 222, 128, 0.1)`
- Warning: `rgba(250, 204, 21, 0.3)` to `rgba(250, 204, 21, 0.15)`
- Error: `rgba(239, 68, 68, 0.3)` to `rgba(239, 68, 68, 0.2)`
- Info: `rgba(96, 165, 250, 0.3)` to `rgba(96, 165, 250, 0.15)`

---

## Component Styles

### Cards

**Base Card**
```css
background: rgba(255, 255, 255, 0.03);
border: 1px solid rgba(255, 255, 255, 0.1);
border-radius: 0.75rem; /* 12px */
padding: 1.5rem; /* 24px */
```

**Glassmorphic Card**
```css
background: rgba(255, 255, 255, 0.05);
backdrop-filter: blur(12px);
border: 1px solid rgba(255, 255, 255, 0.1);
border-radius: 1rem; /* 16px */
```

**Card with Gradient Border**
```css
background: linear-gradient(45deg, rgba(34, 211, 238, 0.5), rgba(16, 185, 129, 0.5)) border-box;
border: 1px solid transparent;
border-radius: 0.75rem;
```

### Buttons

**Primary Button**
```css
background: hsl(var(--primary));
color: #fff;
padding: 0.5rem 1rem; /* 8px 16px */
border-radius: 0.5rem; /* 8px */
font-size: 0.875rem; /* 14px */
font-weight: 500;
```

**Secondary Button**
```css
background: rgba(255, 255, 255, 0.1);
color: rgba(255, 255, 255, 0.9);
border: 1px solid rgba(255, 255, 255, 0.2);
padding: 0.5rem 1rem;
border-radius: 0.5rem;
```

**Ghost Button**
```css
background: transparent;
color: rgba(255, 255, 255, 0.7);
border: 1px solid rgba(255, 255, 255, 0.1);
padding: 0.5rem 1rem;
border-radius: 0.5rem;
```

**Button Hover States**
- Opacity: 0.8 to 0.9
- Background lightness: +5-10%
- Add subtle glow effect

### Navigation Items

**Navigation Link**
```css
color: rgba(255, 255, 255, 0.7);
font-size: 0.875rem; /* 14px */
font-weight: 500;
padding: 0.5rem 1rem;
border-radius: 0.375rem;
transition: all 0.2s;
```

**Active Navigation**
```css
color: #fff;
background: rgba(34, 211, 238, 0.1);
border-left: 2px solid #22d3ee;
```

**Hover State**
```css
color: #fff;
background: rgba(255, 255, 255, 0.05);
```

### Status Badges

**Success Badge**
```css
background: rgba(34, 197, 94, 0.1);
color: #4ade80;
border: 1px solid rgba(74, 222, 128, 0.3);
padding: 0.25rem 0.75rem; /* 4px 12px */
border-radius: 9999px;
font-size: 0.75rem; /* 12px */
font-weight: 500;
```

**Warning Badge**
```css
background: rgba(245, 158, 11, 0.1);
color: #facc15;
border: 1px solid rgba(250, 204, 21, 0.3);
padding: 0.25rem 0.75rem;
border-radius: 9999px;
font-size: 0.75rem;
```

**Error Badge**
```css
background: rgba(239, 68, 68, 0.1);
color: #f87171;
border: 1px solid rgba(248, 113, 113, 0.3);
padding: 0.25rem 0.75rem;
border-radius: 9999px;
font-size: 0.75rem;
```

**Info Badge**
```css
background: rgba(96, 165, 250, 0.1);
color: #60a5fa;
border: 1px solid rgba(96, 165, 250, 0.3);
padding: 0.25rem 0.75rem;
border-radius: 9999px;
font-size: 0.75rem;
```

### Input Fields

**Base Input**
```css
background: rgba(255, 255, 255, 0.05);
border: 1px solid rgba(255, 255, 255, 0.1);
border-radius: 0.5rem; /* 8px */
padding: 0.5rem 1rem;
color: #fff;
font-size: 0.875rem;
```

**Focus State**
```css
border-color: rgba(34, 211, 238, 0.5);
outline: none;
box-shadow: 0 0 0 3px rgba(34, 211, 238, 0.1);
```

### Stat Cards

**Metric Value**
```css
font-size: 2.25rem; /* 36px */
font-weight: 700;
color: #fff;
font-family: 'Space Grotesk', sans-serif;
```

**Metric Label**
```css
font-size: 0.875rem; /* 14px */
font-weight: 500;
color: rgba(255, 255, 255, 0.6);
text-transform: uppercase;
letter-spacing: 0.05em;
```

---

## Background Patterns

### Gradient Backgrounds

**Primary Gradient**
```css
background: linear-gradient(150deg, rgba(96, 165, 250, 0.10) 0%, rgba(34, 211, 238, 0.06) 100%);
```

**Violet Gradient**
```css
background: linear-gradient(150deg, rgba(139, 92, 246, 0.07) 0%, rgba(255, 255, 255, 0.02) 100%);
```

**Purple Gradient**
```css
background: linear-gradient(150deg, rgba(168, 85, 247, 0.08) 0%, rgba(139, 92, 246, 0.04) 100%);
```

**Red Gradient**
```css
background: linear-gradient(150deg, rgba(239, 68, 68, 0.07) 0%, rgba(139, 92, 246, 0.04) 100%);
```

**Ambient Gradient**
```css
background: linear-gradient(90deg, rgba(167, 139, 250, 0.08) 0%, rgba(245, 158, 11, 0.06) 50%, rgba(167, 139, 250, 0.08) 100%);
```

### Radial Gradients

**Header Glow**
```css
background: radial-gradient(circle at 50% 0%, rgba(255, 255, 255, 0.07) 0%, transparent 65%);
```

### Fade Effects

**Left Fade**
```css
background: linear-gradient(to right, #020617 0%, transparent 100%);
```

**Right Fade**
```css
background: linear-gradient(to left, #020617 0%, transparent 100%);
```

### Scanline Pattern

```css
background: repeating-linear-gradient(
  0deg,
  transparent,
  transparent 2px,
  rgba(0, 0, 0, 0.08) 2px,
  rgba(0, 0, 0, 0.08) 4px
);
```

---

## Effects & Animations

### Glassmorphism

**Standard Glass Effect**
```css
background: rgba(255, 255, 255, 0.05);
backdrop-filter: blur(12px);
border: 1px solid rgba(255, 255, 255, 0.1);
```

**Subtle Glass**
```css
background: rgba(255, 255, 255, 0.03);
backdrop-filter: blur(8px);
border: 1px solid rgba(255, 255, 255, 0.05);
```

**Enhanced Glass**
```css
background: rgba(255, 255, 255, 0.1);
backdrop-filter: blur(16px);
border: 1px solid rgba(255, 255, 255, 0.2);
```

### Shadow System

**Subtle Shadow**
```css
box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
```

**Medium Shadow**
```css
box-shadow: 0 4px 6px rgba(0, 0, 0, 0.4);
```

**Large Shadow**
```css
box-shadow: 0 10px 15px rgba(0, 0, 0, 0.5);
```

**Glow Effect (Cyan)**
```css
box-shadow: 0 0 20px rgba(34, 211, 238, 0.3);
```

**Glow Effect (Violet)**
```css
box-shadow: 0 0 20px rgba(167, 139, 250, 0.3);
```

### Transition Timing

**Default Transition**
```css
transition: all 0.2s ease;
```

**Smooth Transition**
```css
transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
```

---

## Layout Patterns

### Container Widths

- Full: `100%`
- Max Content: `max-w-7xl` (1280px)
- Content: `max-w-6xl` (1152px)
- Narrow: `max-w-4xl` (896px)

### Grid System

**Dashboard Grid**
```css
display: grid;
grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
gap: 1.5rem; /* 24px */
```

**Stat Grid**
```css
display: grid;
grid-template-columns: repeat(4, 1fr);
gap: 1rem; /* 16px */
```

### Flexbox Patterns

**Space Between**
```css
display: flex;
justify-content: space-between;
align-items: center;
gap: 1rem;
```

**Center Content**
```css
display: flex;
justify-content: center;
align-items: center;
```

---

## Responsive Breakpoints

| Breakpoint | Min Width | Usage |
|-----------|-----------|-------|
| sm | 640px | Small tablets |
| md | 768px | Tablets |
| lg | 1024px | Desktops |
| xl | 1280px | Large desktops |
| 2xl | 1536px | Extra large screens |

---

## Status Color Coding

### Network Status

- **Online/Active**: Green (#4ade80)
- **Warning/Degraded**: Yellow (#facc15)
- **Error/Down**: Red (#f87171)
- **Info/Processing**: Cyan (#22d3ee)
- **Unknown/Inactive**: Gray (#64748b)

### Alert Severity

- **Critical**: Red (#ef4444)
- **High**: Orange (#fb923c)
- **Medium**: Yellow (#facc15)
- **Low**: Blue (#60a5fa)
- **Info**: Cyan (#22d3ee)

### Performance Indicators

- **Excellent**: Green (#22c55e)
- **Good**: Teal (#2dd4bf)
- **Fair**: Yellow (#facc15)
- **Poor**: Orange (#f97316)
- **Critical**: Red (#ef4444)

---

## CSS Variables (HSL Format)

```css
:root {
  --background: 222.2 84% 4.9%;
  --foreground: 0 0% 100%;
  --primary: 189 95% 52%; /* Cyan */
  --muted-foreground: 215 16.3% 56.9%;
  --ring: 189 95% 52%;
}
```

---

## Icon Sizing

| Size | Dimensions | Usage |
|------|-----------|--------|
| xs | 12px | Inline icons |
| sm | 16px | Small UI elements |
| base | 20px | Default icons |
| md | 24px | Card headers |
| lg | 32px | Feature icons |
| xl | 48px | Hero sections |
| 2xl | 64px | Large displays |

---

## Z-Index Scale

| Layer | Value | Usage |
|-------|-------|-------|
| Base | 0 | Default content |
| Dropdown | 10 | Dropdown menus |
| Sticky | 20 | Sticky headers |
| Fixed | 30 | Fixed elements |
| Modal Backdrop | 40 | Modal overlays |
| Modal | 50 | Modal dialogs |
| Popover | 60 | Popovers & tooltips |
| Toast | 70 | Toast notifications |
| Tooltip | 80 | Tooltips |

---

## Best Practices

### Glassmorphism Usage

1. Use `rgba(255, 255, 255, 0.03)` to `0.1` for card backgrounds
2. Apply `backdrop-filter: blur(8-16px)` for glass effect
3. Combine with subtle borders: `rgba(255, 255, 255, 0.1)`
4. Layer multiple glass elements carefully to maintain readability

### Color Contrast

1. Maintain minimum contrast ratio of 4.5:1 for text
2. Use white text on dark backgrounds
3. Apply opacity to reduce visual weight: 0.6-0.9 for secondary text
4. Test readability with glassmorphic overlays

### Animation Performance

1. Prefer `opacity` and `transform` for animations
2. Use `transition` for interactive states
3. Keep transition durations between 0.2s-0.3s
4. Apply `will-change` sparingly for performance-critical animations

### Spacing Consistency

1. Use 8px base unit for spacing (0.5rem)
2. Maintain consistent padding in similar components
3. Use larger spacing for visual hierarchy
4. Apply responsive spacing adjustments at breakpoints

---

## File References

This style guide was extracted from the following HTML files:
- `NOC System - Enterprise Network Management by IndoBSD.html`
- `NOC System - Enterprise Network Management by IndoBSD (15_09_2026 02：53：30).html`
- `NOC System - Enterprise Network Management by IndoBSD (15_09_2026 06：51：28).html`
- `NOC System - Enterprise Network Management by IndoBSD (15_09_2026 06：51：51).html`
- `NOC System - Enterprise Network Management by IndoBSD (15_09_2026 06：52：31).html`
- `NOC System - Enterprise Network Management by IndoBSD (15_09_2026 06：52：53).html`

---

## Version

**Document Version**: 1.0  
**Last Updated**: 2026-09-15  
**Design System**: NOC Enterprise Network Management  
**Created By**: IndoBSD

---

*This comprehensive style guide ensures consistency across all NOC system interfaces and provides clear specifications for implementation.*
