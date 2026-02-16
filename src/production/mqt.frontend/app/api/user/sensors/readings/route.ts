import { NextRequest, NextResponse } from "next/server";
import { READINGS_API_BASE_URL } from "@/constants/api";

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const deviceId = searchParams.get("device_id");
    const piId = searchParams.get("pi_id");
    const page = searchParams.get("page") || "1";
    const limit = searchParams.get("limit") || "20";
    const from = searchParams.get("from");
    const to = searchParams.get("to");

    // Get authentication token from Authorization header
    const authHeader = request.headers.get("authorization");
    if (!authHeader || !authHeader.startsWith("Bearer ")) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    if (!piId) {
      return NextResponse.json({ error: "pi_id is required" }, { status: 400 });
    }

    if (!deviceId) {
      return NextResponse.json({ error: "device_id is required" }, { status: 400 });
    }

    // Build query parameters
    const queryParams = new URLSearchParams();
    if (page) queryParams.append("page", page);
    if (limit) queryParams.append("limit", limit);
    if (from) queryParams.append("from", from);
    if (to) queryParams.append("to", to);

    // URL encode deviceId for the external API call (MAC address with colons)
    const encodedDeviceId = encodeURIComponent(deviceId);
    
    // Call the readings API directly - device_id is MAC address string
    const url = `${READINGS_API_BASE_URL}/readings/pis/${piId}/devices/${encodedDeviceId}${queryParams.toString() ? `?${queryParams}` : ""}`;
    
    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Authorization": authHeader,
        "Content-Type": "application/json",
      },
      credentials: "include",
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to fetch readings" }));
      return NextResponse.json(
        { error: error.error || error.message || "Failed to fetch readings" },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error fetching readings:", error);
    const errorMessage = error instanceof Error ? error.message : "Failed to fetch readings";
    return NextResponse.json({ error: errorMessage }, { status: 500 });
  }
}

