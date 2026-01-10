"use client";

import React from "react";
import { GoogleMap, useJsApiLoader, Marker } from "@react-google-maps/api";

const containerStyle = {
    width: "100%",
    height: "100%",
};



// Default center (Tuscaloosa Default)
const defaultCenter = { lat: 33.1956, lng: -87.5268 };

export default function LiveMap({ pickup, dropoff, shopper }: LiveMapProps) {
    // Local defaultCenter removed to prevent reference churn

    // Determine center: Pickup > Dropoff > Default
    // Do NOT use 'shopper' here, as it changes frequently and causes the map to jump.
    // relying on fitBounds vs user interaction for actual view.
    const center = pickup || dropoff || defaultCenter;

    const { isLoaded } = useJsApiLoader({
        id: "google-map-script",
        googleMapsApiKey: process.env.NEXT_PUBLIC_GOOGLE_MAPS_KEY || "",
    });

    const [map, setMap] = React.useState<google.maps.Map | null>(null);

    const onLoad = React.useCallback(function callback(map: google.maps.Map) {
        setMap(map);
    }, []);

    const onUnmount = React.useCallback(function callback() {
        setMap(null);
    }, []);

    const hasFitBounds = React.useRef(false);



    const fitMapToBounds = React.useCallback(() => {
        if (!map) return;

        const bounds = new google.maps.LatLngBounds();
        let pointCount = 0;

        if (shopper) { bounds.extend(shopper); pointCount++; }
        if (pickup) { bounds.extend(pickup); pointCount++; }
        if (dropoff) { bounds.extend(dropoff); pointCount++; }

        if (pointCount > 0) {
            map.fitBounds(bounds);

            // If only one point, avoid max zoom
            if (pointCount === 1) {
                const listener = google.maps.event.addListener(map, "idle", () => {
                    if (map.getZoom()! > 15) map.setZoom(15);
                    google.maps.event.removeListener(listener);
                });
            }
        } else {
            map.panTo(defaultCenter);
            map.setZoom(14);
        }
    }, [map, shopper, pickup, dropoff]);

    // Auto-fit bounds ONLY on initial load of points
    React.useEffect(() => {
        if (map && !hasFitBounds.current && (pickup || dropoff)) {
            fitMapToBounds();
            hasFitBounds.current = true;
        }
    }, [map, pickup, dropoff, fitMapToBounds]); // Dependent on fitMapToBounds

    if (!isLoaded) {
        return (
            <div className="w-full h-full flex items-center justify-center bg-zinc-100 dark:bg-zinc-800 text-zinc-500">
                Loading Maps...
            </div>
        );
    }

    return (
        <div className="relative w-full h-full">
            <GoogleMap
                mapContainerStyle={containerStyle}
                center={defaultCenter} // Use STATIC center to make map uncontrolled.
                zoom={14}
                onLoad={onLoad}
                onUnmount={onUnmount}
                options={{
                    disableDefaultUI: true, // We will build our own controls eventually?
                    zoomControl: false,      // Keep default zoom hidden if we want custom look, or true. let's keep false for clean look if we have manual center.
                    // Actually user asked for a button. Custom UI is better.
                }}
            >
                {/* Shopper (Blue Arrow/Car) */}
                {shopper && (
                    <Marker
                        position={shopper}
                        title="Me"
                        icon={{
                            url: "https://maps.google.com/mapfiles/kml/shapes/cabs.png",
                            scaledSize: new google.maps.Size(30, 30)
                        }}
                    />
                )}

                {/* Pickup/Store (Restaurant/Store Icon) */}
                {pickup && (
                    <Marker
                        position={pickup}
                        title="Pickup - Target"
                        icon={{
                            url: "https://maps.google.com/mapfiles/kml/shapes/shopping.png", // Or a better restaurant icon
                            scaledSize: new google.maps.Size(40, 40)
                        }}
                    />
                )}

                {/* Dropoff/Customer (House/Pin) */}
                {dropoff && (
                    <Marker
                        position={dropoff}
                        title="Dropoff - Customer"
                        animation={google.maps.Animation.BOUNCE}
                        icon={{
                            url: "https://maps.google.com/mapfiles/kml/shapes/homegardenbusiness.png", // House icon
                            scaledSize: new google.maps.Size(40, 40)
                        }}
                    />
                )}
            </GoogleMap>

            {/* Re-Center Button */}
            <button
                onClick={fitMapToBounds}
                className="absolute top-4 right-4 bg-white dark:bg-zinc-800 text-zinc-700 dark:text-zinc-200 p-3 rounded-full shadow-lg hover:bg-zinc-50 dark:hover:bg-zinc-700 transition-colors z-10"
                title="Fit to Bounds"
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="15 3 21 3 21 9"></polyline>
                    <polyline points="9 21 3 21 3 15"></polyline>
                    <line x1="21" y1="3" x2="14" y2="10"></line>
                    <line x1="3" y1="21" x2="10" y2="14"></line>
                </svg>
            </button>
        </div>
    );
}
