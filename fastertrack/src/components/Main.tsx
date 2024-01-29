import { useState } from "react";

import Stack from "@mui/joy/Stack";

import Runs from "./Runs";
import SearchRuns from "./SearchRuns";
import Metrics from "./Metrics";
import SelectMetrics from "./SelectMetrics";

export default function Main({ selectedExperiments }: { selectedExperiments: string[] }) {
    const [searchQuery, setSearchQuery] = useState('' as string);
    const [selectedRuns, setSelectedRuns] = useState([] as string[]);
    const [availableMetrics, setAvailableMetrics] = useState([] as string[]);
    const [selectedMetrics, setSelectedMetrics] = useState<string[]>([]);

    return (
        <Stack
            direction='column'
            justifyContent='center'
            alignItems='center'
            spacing={2}
            sx={{
                height: '100%',
            }}
        >
            <Stack
                direction='row'
                justifyContent='center'
                alignItems='center'
                spacing={2}
                sx={{
                    width: '100%',
                }}
            >
                <SearchRuns setSearchQuery={setSearchQuery} sx={{ width: '50%' }} />
                <SelectMetrics availableMetrics={availableMetrics} setSelectedMetrics={setSelectedMetrics} sx={{ width: '50%' }} />
            </Stack>
            <Metrics selectedMetrics={selectedMetrics} selectedRuns={selectedRuns} />
            <Runs selectedExperiments={selectedExperiments} searchQuery={searchQuery} setAvailableMetrics={setAvailableMetrics} setSelectedRuns={setSelectedRuns} />
        </Stack >
    )
}