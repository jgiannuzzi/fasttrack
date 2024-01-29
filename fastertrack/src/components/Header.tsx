import Box from '@mui/joy/Box';
import Typography from '@mui/joy/Typography';
import Stack from '@mui/joy/Stack';

export default function Header() {
    return (
        <Box
            sx={{
                display: 'flex',
                flexGrow: 1,
                justifyContent: 'space-between',
            }}
        >
            <Stack
                direction="row"
                justifyContent="center"
                alignItems="center"
                spacing={1}
            >
                <Box></Box>
                <Typography level="h1">
                    <em>FasterTrack</em>
                </Typography>
            </Stack>
        </Box >
    );
}