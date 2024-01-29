import { Dispatch, SetStateAction, KeyboardEvent } from 'react';

import Box from '@mui/joy/Box';
import Input from '@mui/joy/Input';
import { SxProps } from '@mui/joy/styles/types';

export default function SearchRuns({
    setSearchQuery,
    sx,
}: {
    setSearchQuery: Dispatch<SetStateAction<string>>
    sx?: SxProps
}) {
    const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === 'Enter') {
            setSearchQuery(event.currentTarget.value);
        }
    };

    return (
        <Box
            sx={sx}
        >
            <Input
                size="sm"
                variant="outlined"
                placeholder="Search for runs…"
                onKeyDown={onKeyDown}
                slotProps={{
                    input: {
                        spellCheck: 'false'
                    }
                }}
            />
        </Box>
    )
}