import { useState, useEffect } from "react";

interface Location {
    lat: number;
    lng: number;
}

export function useGeolocation(enabled: boolean = true) {
    const [location, setLocation] = useState<Location | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!enabled || !("geolocation" in navigator)) {
            setError("Geolocation is not supported or disabled");
            return;
        }

        const handleSuccess = (position: GeolocationPosition) => {
            setLocation({
                lat: position.coords.latitude,
                lng: position.coords.longitude,
            });
        };

        const handleError = (error: GeolocationPositionError) => {
            setError(error.message);
        };

        const watcher = navigator.geolocation.watchPosition(handleSuccess, handleError, {
            enableHighAccuracy: true,
            timeout: 5000,
            maximumAge: 0,
        });

        return () => navigator.geolocation.clearWatch(watcher);
    }, [enabled]);

    return { location, error };
}
