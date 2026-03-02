import { useState, useEffect, useCallback } from "react";
import { deviceLocationService } from "@/services/api/deviceLocationService";
import { tailwindToHex, getMostCommonColor } from "@/lib/colorUtils";

export interface PiPreference {
  pi_id: string;
  color: string; // Hex color code (e.g., "#f97316")
}

const STORAGE_KEY = "pi_marker_preferences";

// Default colors (no blue - blue is reserved for current location marker)
const DEFAULT_COLORS = [
  "#f97316", // orange-500
  "#22c55e", // green-500
  "#a855f7", // purple-500
  "#ef4444", // red-500
  "#eab308", // yellow-500
  "#ec4899", // pink-500
  "#6366f1", // indigo-500
  "#14b8a6", // teal-500
  "#a855f7", // violet-500
  "#10b981", // emerald-500
];

/**
 * Convert a hex color to a gradient string for display purposes
 */
export function colorToGradient(color: string): string {
  if (!color) {
    return `from-[#f97316] to-[#f97316]`;
  }
  // Return gradient with same color for both from and to (MarkerShape will handle darkening)
  return `from-[${color}] to-[${color}]`;
}

export function usePiPreferences() {
  const [preferences, setPreferences] = useState<Record<string, PiPreference>>({});
  const [isLoadingFromBackend, setIsLoadingFromBackend] = useState(false);

  // Load preferences from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        // Validate and clean preferences (ensure no null/undefined colors)
        const cleaned: Record<string, PiPreference> = {};
        Object.entries(parsed).forEach(([pi_id, pref]: [string, unknown]) => {
          const p = pref as { color?: string; gradient?: string } | null | undefined;
          if (p && typeof p === 'object') {
            let color: string;
            
            // Handle migration from old gradient format to new color format
            if (p.color) {
              // Already in new format
              color = p.color;
            } else if (p.gradient) {
              // Migrate from gradient format
              // Check for custom hex colors first: from-[#f97316]
              const hexFromMatch = p.gradient.match(/from-\[#([0-9A-Fa-f]{6})\]/);
              if (hexFromMatch) {
                color = `#${hexFromMatch[1]}`;
              } else {
                // Try Tailwind class: from-orange-500
                const fromMatch = p.gradient.match(/from-([a-z]+-\d+)/);
                if (fromMatch) {
                  color = tailwindToHex(fromMatch[1]);
                } else {
                  color = DEFAULT_COLORS[0];
                }
              }
            } else {
              color = DEFAULT_COLORS[0];
            }
            
            cleaned[pi_id] = {
              pi_id: pi_id,
              color: color,
            };
          }
        });
        setPreferences(cleaned);
        // Save migrated data back to localStorage
        if (Object.keys(cleaned).length > 0) {
          localStorage.setItem(STORAGE_KEY, JSON.stringify(cleaned));
        }
      }
    } catch (error) {
      console.error("Error loading PI preferences:", error);
      // Clear corrupted localStorage data
      try {
        localStorage.removeItem(STORAGE_KEY);
      } catch {
        // Ignore errors when clearing
      }
    }
  }, []);

  // Load colors from backend device locations
  const loadColorsFromBackend = useCallback(async () => {
    try {
      setIsLoadingFromBackend(true);
      const locations = await deviceLocationService.getAllLocations();
      
      // Group locations by PI and get most common color for each PI
      const piColorMap: Record<string, string[]> = {};
      const piSet = new Set<string>(); // Track all PIs found in locations
      
      locations.forEach(location => {
        piSet.add(location.pi_id);
        if (location.color) {
          if (!piColorMap[location.pi_id]) {
            piColorMap[location.pi_id] = [];
          }
          piColorMap[location.pi_id].push(location.color);
        }
      });

      // Update preferences with colors from backend
      setPreferences(currentPreferences => {
        const updated = { ...currentPreferences };
        let hasChanges = false;

        // Process PIs that have colors in backend
        Object.entries(piColorMap).forEach(([pi_id, colors]) => {
          const mostCommonHex = getMostCommonColor(colors);
          if (mostCommonHex) {
            const currentPref = updated[pi_id];
            
            // Only update if color changed or preference doesn't exist
            if (!currentPref || currentPref.color !== mostCommonHex) {
              hasChanges = true;
              updated[pi_id] = {
                pi_id,
                color: mostCommonHex,
              };
            }
          }
        });

        // Ensure all PIs have preferences (assign defaults for PIs without colors)
        Array.from(piSet).forEach((pi_id, index) => {
          if (!updated[pi_id]) {
            hasChanges = true;
            const colorIndex = index % DEFAULT_COLORS.length;
            updated[pi_id] = {
              pi_id,
              color: DEFAULT_COLORS[colorIndex],
            };
          } else if (!updated[pi_id].color) {
            // If preference exists but has no color, assign default
            hasChanges = true;
            const colorIndex = index % DEFAULT_COLORS.length;
            updated[pi_id] = {
              ...updated[pi_id],
              color: DEFAULT_COLORS[colorIndex],
            };
          }
        });

        if (hasChanges) {
          try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
          } catch (error) {
            console.error("Error saving preferences to localStorage:", error);
          }
        }

        return updated;
      });
    } catch (error) {
      console.error("Error loading colors from backend:", error);
      // Don't throw - fall back to localStorage preferences
    } finally {
      setIsLoadingFromBackend(false);
    }
  }, []);

  // Save preferences to localStorage
  const savePreferences = useCallback((newPreferences: Record<string, PiPreference>) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newPreferences));
      setPreferences(newPreferences);
    } catch (error) {
      console.error("Error saving PI preferences:", error);
    }
  }, []);

  // Get preference for a specific PI, with auto-assignment if not exists
  const getPreference = useCallback(
    (pi_id: string, index: number = 0): PiPreference => {
      const existing = preferences[pi_id];
      
      // If preference exists and has valid color, return it
      if (existing?.color) {
        return existing;
      }

      // Auto-assign based on index (for new PIs or invalid preferences)
      const colorIndex = index % DEFAULT_COLORS.length;
      
      const defaultPreference: PiPreference = {
        pi_id,
        color: DEFAULT_COLORS[colorIndex],
      };

      // If preference exists but is missing color, merge with defaults
      if (existing) {
        return {
          ...defaultPreference,
          ...existing,
          color: existing.color || defaultPreference.color,
        };
      }

      return defaultPreference;
    },
    [preferences]
  );

  // Update preference for a specific PI
  const updatePreference = useCallback(
    async (pi_id: string, updates: Partial<Omit<PiPreference, "pi_id">>) => {
      const current = preferences[pi_id] || getPreference(pi_id);
      
      // Ensure color is always valid (use defaults if null/undefined)
      const safeUpdates: Partial<Omit<PiPreference, "pi_id">> = {
        ...updates,
        color: updates.color || current.color || DEFAULT_COLORS[0],
      };
      
      const updated = { ...preferences, [pi_id]: { ...current, ...safeUpdates } };
      
      // Save to localStorage immediately
      savePreferences(updated);

      // If color is being updated, sync to backend
      if (safeUpdates.color) {
        try {
          await deviceLocationService.updatePiColor(pi_id, safeUpdates.color);
        } catch (error) {
          console.error(`Error syncing color to backend for PI ${pi_id}:`, error);
          // Don't throw - localStorage is already updated, backend sync can retry later
        }
      }
    },
    [preferences, getPreference, savePreferences]
  );

  // Initialize preferences for multiple PIs
  const initializePreferences = useCallback(
    (pi_ids: string[]) => {
      setPreferences((currentPreferences) => {
        const newPreferences: Record<string, PiPreference> = { ...currentPreferences };
        let hasChanges = false;
        
        pi_ids.forEach((pi_id, index) => {
          if (!newPreferences[pi_id]) {
            hasChanges = true;
            const colorIndex = index % DEFAULT_COLORS.length;
            newPreferences[pi_id] = {
              pi_id,
              color: DEFAULT_COLORS[colorIndex],
            };
          }
        });
        
        // Only save to localStorage if there are changes
        if (hasChanges) {
          try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(newPreferences));
          } catch (error) {
            console.error("Error saving PI preferences:", error);
          }
        }
        
        return newPreferences;
      });
    },
    [] // No dependencies needed since we use functional update
  );

  // Reset all preferences
  const resetPreferences = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY);
    setPreferences({});
  }, []);

  return {
    preferences,
    getPreference,
    updatePreference,
    initializePreferences,
    resetPreferences,
    loadColorsFromBackend,
    isLoadingFromBackend,
  };
}

