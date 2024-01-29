import { useEffect } from "react";
import { useColorScheme } from '@mui/joy/styles';

export default function ColorSchemeDetector() {
    const { setMode } = useColorScheme();

    useEffect(() => {
        window.matchMedia('(prefers-color-scheme: dark)')
            .addEventListener('change', event => {
                setMode(event.matches ? "dark" : "light");
            });
    }, [setMode]);

    return (
        <></>
    );
}