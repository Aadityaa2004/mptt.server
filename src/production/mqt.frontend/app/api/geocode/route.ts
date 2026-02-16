import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const query = searchParams.get("q");
    const limit = searchParams.get("limit") || "5";

    if (!query || !query.trim()) {
      return NextResponse.json({ error: "Query parameter 'q' is required" }, { status: 400 });
    }

    // Proxy request to OpenStreetMap Nominatim API
    // This avoids CORS issues and complies with Nominatim's usage policy
    const nominatimUrl = `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=${limit}`;
    
    const response = await fetch(nominatimUrl, {
      headers: {
        "User-Agent": "MapleSense Weather App (https://github.com/yourusername/mptt.ui)", // Nominatim requires a proper User-Agent
        "Accept-Language": "en-US,en;q=0.9", // Optional: specify preferred language
      },
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error("Nominatim API error:", response.status, errorText);
      return NextResponse.json(
        { error: `Geocoding service error: ${response.statusText}` },
        { status: response.status }
      );
    }

    const data = await response.json();
    
    // Transform the data to match the expected format
    const results = data.map((item: { lat: string; lon: string; display_name: string }) => ({
      lat: item.lat,
      lon: item.lon,
      display_name: item.display_name,
    }));

    return NextResponse.json(results);
  } catch (error) {
    console.error("Geocoding proxy error:", error);
    const errorMessage = error instanceof Error ? error.message : "Failed to fetch geocoding data";
    return NextResponse.json({ error: errorMessage }, { status: 500 });
  }
}

