/* eslint-disable react-refresh/only-export-components */
import Box, { BoxProps } from '@mui/joy/Box';

function Root(props: BoxProps) {
    return (
        <Box
            {...props}
            sx={[
                {
                    display: 'grid',
                    gridTemplateColumns: {
                        xs: '250px 1fr',
                    },
                    gridTemplateRows: '64px 1fr',
                    minHeight: '100vh',
                },
                ...(Array.isArray(props.sx) ? props.sx : [props.sx]),
            ]}
        />
    );
}

function Header(props: BoxProps) {
    return (
        <Box
            component="header"
            className="Header"
            {...props}
            sx={[
                {
                    p: 2,
                    gap: 2,
                    bgcolor: 'background.surface',
                    display: 'flex',
                    flexDirection: 'row',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    gridColumn: '1 / -1',
                    borderBottom: '1px solid',
                    borderColor: 'divider',
                    position: 'sticky',
                    top: 0,
                    zIndex: 1100,
                },
                ...(Array.isArray(props.sx) ? props.sx : [props.sx]),
            ]}
        />
    );
}

function SideNav(props: BoxProps) {
    return (
        <Box
            component="nav"
            className="Navigation"
            {...props}
            sx={[
                {
                    p: 2,
                    bgcolor: 'background.surface',
                    borderRight: '1px solid',
                    borderColor: 'divider',
                },
                ...(Array.isArray(props.sx) ? props.sx : [props.sx]),
            ]}
        />
    );
}

function Main(props: BoxProps) {
    return (
        <Box
            component="main"
            className="Main"
            {...props}
            sx={[{ p: 2 }, ...(Array.isArray(props.sx) ? props.sx : [props.sx])]}
        />
    );
}



export default {
    Root,
    Header,
    SideNav,
    Main,
};