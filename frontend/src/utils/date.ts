
export const parseDate = (time: string | number | undefined | null): Date => {
    if (!time) return new Date(0) // Return epoch if empty

    // If number, direct date
    if (typeof time === 'number') {
        return new Date(time)
    }

    // If string, check if numeric
    const cleanTime = String(time).trim()
    if (/^\d+(\.\d+)?$/.test(cleanTime)) {
        return new Date(parseInt(cleanTime))
    }

    // Else parse as string
    return new Date(cleanTime)
}
