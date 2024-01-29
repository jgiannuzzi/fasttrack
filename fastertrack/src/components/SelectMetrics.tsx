import { Fragment, SyntheticEvent, Dispatch, SetStateAction } from "react";

import Box from "@mui/joy/Box";
import Chip from "@mui/joy/Chip";
import Option from "@mui/joy/Option";
import Select from "@mui/joy/Select";
import { SxProps } from "@mui/joy/styles/types";

export default function SelectMetrics({
    availableMetrics,
    setSelectedMetrics,
    sx,
}: {
    availableMetrics: string[],
    setSelectedMetrics: Dispatch<SetStateAction<string[]>>,
    sx?: SxProps
}) {
    const onChange = (
        _event: SyntheticEvent | null,
        value: Array<string> | null,
    ) => {
        setSelectedMetrics(value || []);
    };

    return (
        <Select
            placeholder="Select metrics"
            multiple
            renderValue={(selected) => (
                <Box sx={{ display: 'flex', gap: '0.25rem' }}>
                    {selected.map((selectedOption) => (
                        <Chip key={selectedOption.id} variant="soft" color="primary">
                            {selectedOption.label}
                        </Chip>
                    ))}
                </Box>
            )}
            onChange={onChange}
            sx={sx}
        >
            {availableMetrics.map((metric) => {
                return (
                    <Fragment key={metric}>
                        <Option value={metric}>{metric}</Option>
                    </Fragment>
                )
            })}
        </Select>
    )
}