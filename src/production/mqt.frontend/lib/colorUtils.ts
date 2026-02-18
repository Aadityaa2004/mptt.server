// Color conversion utilities for Tailwind classes and hex codes

// Map Tailwind color classes to hex codes
const TAILWIND_TO_HEX: Record<string, string> = {
  "orange-500": "#f97316",
  "blue-500": "#3b82f6",
  "green-500": "#22c55e",
  "purple-500": "#a855f7",
  "red-500": "#ef4444",
  "yellow-500": "#eab308",
  "pink-500": "#ec4899",
  "cyan-500": "#06b6d4",
  "indigo-500": "#6366f1",
  "teal-500": "#14b8a6",
};

// Map hex codes to Tailwind color classes
const HEX_TO_TAILWIND: Record<string, string> = Object.fromEntries(
  Object.entries(TAILWIND_TO_HEX).map(([tailwind, hex]) => [hex.toLowerCase(), tailwind])
);

/**
 * Convert Tailwind color class to hex code
 * Returns default hex if input is null/undefined or invalid
 */
export function tailwindToHex(tailwindColor: string | null | undefined): string {
  if (!tailwindColor) {
    return TAILWIND_TO_HEX["orange-500"]; // Default color
  }
  return TAILWIND_TO_HEX[tailwindColor] || TAILWIND_TO_HEX["orange-500"];
}

/**
 * Convert hex code to Tailwind color class (if it matches)
 * Returns default Tailwind color if no match is found or if input is null/undefined
 */
export function hexToTailwind(hexColor: string | null | undefined): string {
  if (!hexColor) {
    return "orange-500"; // Default color
  }
  const normalizedHex = hexColor.toLowerCase();
  return HEX_TO_TAILWIND[normalizedHex] || "orange-500"; // Default to orange-500 if no match
}

/**
 * Get the most common color from an array of hex colors
 */
export function getMostCommonColor(colors: (string | undefined)[]): string | undefined {
  const validColors = colors.filter((c): c is string => Boolean(c));
  if (validColors.length === 0) return undefined;

  const colorCounts: Record<string, number> = {};
  validColors.forEach(color => {
    colorCounts[color] = (colorCounts[color] || 0) + 1;
  });

  let maxCount = 0;
  let mostCommon: string | undefined;
  Object.entries(colorCounts).forEach(([color, count]) => {
    if (count > maxCount) {
      maxCount = count;
      mostCommon = color;
    }
  });

  return mostCommon;
}

/**
 * Extract primary color from gradient string (from-{color} to-{color})
 * Returns the "from" color
 */
export function gradientToHex(gradient: string): string {
  if (!gradient) {
    return TAILWIND_TO_HEX["orange-500"];
  }
  
  const match = gradient.match(/from-([a-z]+-\d+)/);
  if (match && match[1]) {
    return tailwindToHex(match[1]);
  }
  
  return TAILWIND_TO_HEX["orange-500"];
}

