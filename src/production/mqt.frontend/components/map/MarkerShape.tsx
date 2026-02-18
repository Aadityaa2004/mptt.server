import { tailwindToHex } from "@/lib/colorUtils";

interface MarkerComponentProps {
  gradient: string; // Gradient string format: "from-blue-500 to-purple-500" or "from-[#f97316] to-[#f97316]"
  size?: "sm" | "md" | "lg";
  isSelected?: boolean;
}

// Helper functions for color manipulation
function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result ? {
    r: parseInt(result[1], 16),
    g: parseInt(result[2], 16),
    b: parseInt(result[3], 16)
  } : null;
}

function rgbToHex(r: number, g: number, b: number): string {
  return "#" + [r, g, b].map(x => {
    const hex = x.toString(16);
    return hex.length === 1 ? "0" + hex : hex;
  }).join("");
}

// Extract colors from gradient string and convert to hex
function parseGradient(gradient: string): { from: string; to: string; border: string; shadow: string } {
  // Handle custom hex colors: from-[#f97316] to-[#ef4444]
  const hexFromMatch = gradient.match(/from-\[#([0-9A-Fa-f]{6})\]/);
  const hexToMatch = gradient.match(/to-\[#([0-9A-Fa-f]{6})\]/);
  
  let fromHex: string;
  let toHex: string;
  
  if (hexFromMatch) {
    // Custom hex color
    fromHex = `#${hexFromMatch[1]}`;
  } else {
    // Tailwind class
    const fromMatch = gradient.match(/from-([a-z]+-\d+)/);
    const fromColor = fromMatch ? fromMatch[1] : "orange-500";
    fromHex = tailwindToHex(fromColor);
  }
  
  if (hexToMatch) {
    // Custom hex color
    toHex = `#${hexToMatch[1]}`;
  } else {
    // Tailwind class
    const toMatch = gradient.match(/to-([a-z]+-\d+)/);
    const toColor = toMatch ? toMatch[1] : "orange-600";
    toHex = tailwindToHex(toColor);
  }
  
  // If both colors are the same (single color), create a subtle gradient by darkening "to" slightly
  if (fromHex === toHex) {
    // Convert hex to RGB, darken by 10%, convert back to hex
    const rgb = hexToRgb(fromHex);
    if (rgb) {
      const darkened = {
        r: Math.max(0, rgb.r - 20),
        g: Math.max(0, rgb.g - 20),
        b: Math.max(0, rgb.b - 20),
      };
      toHex = rgbToHex(darkened.r, darkened.g, darkened.b);
    }
  }
  
  // Create border color (lighter version of from color)
  const borderHex = fromHex; // Use same as from for simplicity
  const shadowHex = fromHex + "B3"; // Add opacity
  
  return {
    from: fromHex,
    to: toHex,
    border: borderHex,
    shadow: shadowHex,
  };
}

export function MarkerShapeComponent({
  gradient,
  size = "md",
  isSelected = false,
}: MarkerComponentProps) {
  const colorValues = parseGradient(gradient);
  
  const sizeValue = size === "sm" ? 20 : size === "md" ? 24 : 32;
  const innerSizeValue = size === "sm" ? 8 : size === "md" ? 10 : 12;
  const borderWidth = 2;

  const baseStyle: React.CSSProperties = {
    width: `${sizeValue}px`,
    height: `${sizeValue}px`,
    background: `linear-gradient(135deg, ${colorValues.from}, ${colorValues.to})`,
    border: `${borderWidth}px solid ${colorValues.border}`,
    boxShadow: `0 4px 6px -1px ${colorValues.shadow}`,
    borderRadius: "50%", // Always circle
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    position: "relative",
    transition: "all 0.2s",
    transform: isSelected ? "scale(1.1)" : "scale(1)",
    filter: isSelected ? `drop-shadow(0 0 8px ${colorValues.shadow})` : undefined,
  };

  const innerStyle: React.CSSProperties = {
    width: `${innerSizeValue}px`,
    height: `${innerSizeValue}px`,
    backgroundColor: "white",
    borderRadius: "50%",
    position: "absolute",
  };

  return (
    <div
      className="cursor-pointer transition-all hover:scale-110"
      style={baseStyle}
    >
      <div style={innerStyle} />
    </div>
  );
}

