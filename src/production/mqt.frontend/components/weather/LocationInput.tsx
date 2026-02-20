"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { MapPin, Loader2 } from "lucide-react";

interface LocationInputProps {
  onLocationSubmit: (latitude: number, longitude: number, locationName: string) => void;
  isLoading?: boolean;
}

interface GeocodeResult {
  lat: string;
  lon: string;
  display_name: string;
}

export function LocationInput({ onLocationSubmit, isLoading }: LocationInputProps) {
  const [query, setQuery] = useState("");
  const [isSearching, setIsSearching] = useState(false);
  const [suggestions, setSuggestions] = useState<GeocodeResult[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const searchLocation = async (searchQuery: string) => {
    if (!searchQuery.trim()) {
      setSuggestions([]);
      setShowSuggestions(false);
      return;
    }

    setIsSearching(true);
    try {
      // Use Next.js API route to proxy the request to Nominatim
      // This avoids CORS issues and complies with Nominatim's usage policy
      const response = await fetch(
        `/api/geocode?q=${encodeURIComponent(searchQuery)}&limit=5`
      );

      if (response.ok) {
        const data = await response.json();
        setSuggestions(data);
        setShowSuggestions(true);
      } else {
        const errorData = await response.json().catch(() => ({ error: "Failed to fetch locations" }));
        console.error("Geocoding error:", errorData.error || response.statusText);
        setSuggestions([]);
        setShowSuggestions(false);
      }
    } catch (error) {
      console.error("Geocoding error:", error);
      setSuggestions([]);
      setShowSuggestions(false);
    } finally {
      setIsSearching(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    if (value.length > 2) {
      searchLocation(value);
    } else {
      setSuggestions([]);
      setShowSuggestions(false);
    }
  };

  const handleSuggestionClick = (suggestion: GeocodeResult) => {
    setQuery(suggestion.display_name);
    setShowSuggestions(false);
    onLocationSubmit(
      parseFloat(suggestion.lat),
      parseFloat(suggestion.lon),
      suggestion.display_name
    );
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (suggestions.length > 0) {
      handleSuggestionClick(suggestions[0]);
    }
  };

  return (
    <div className="relative w-full max-w-xl">
      <form onSubmit={handleSubmit} className="flex gap-2">
        <div className="relative flex-1">
          <MapPin className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-white/40" />
          <Input
            type="text"
            value={query}
            onChange={handleInputChange}
            placeholder="Search for a city or location..."
            className="pl-10 h-11 rounded-xl bg-white/[0.06] border-white/10 text-white placeholder:text-white/40 focus-visible:ring-1 focus-visible:ring-orange-500/30 focus-visible:border-orange-500/30"
            disabled={isLoading}
          />
          {isSearching && (
            <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-white/40 animate-spin" />
          )}
        </div>
        <Button
          type="submit"
          disabled={isLoading || !query.trim()}
          className="h-11 px-5 rounded-xl bg-orange-500/90 hover:bg-orange-500 text-white font-light transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Saving...
            </>
          ) : (
            "Set Location"
          )}
        </Button>
      </form>

      {showSuggestions && suggestions.length > 0 && (
        <div className="absolute z-50 w-full mt-2 rounded-xl overflow-hidden bg-black/95 border border-white/10 shadow-xl">
          {suggestions.map((suggestion, index) => (
            <button
              key={index}
              type="button"
              onClick={() => handleSuggestionClick(suggestion)}
              className="w-full px-4 py-3 text-left hover:bg-white/5 active:bg-white/10 transition-colors flex items-center gap-3 border-b border-white/5 last:border-b-0 first:pt-3 last:pb-3"
            >
              <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center flex-shrink-0">
                <MapPin className="h-4 w-4 text-orange-400/80" />
              </div>
              <span className="text-sm text-white/90 font-light truncate">
                {suggestion.display_name}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

