<template>
    <div class="contribution-graph select-none w-full">
        <!-- Header: Month Navigation -->
        <div class="flex items-center justify-between mb-4">
            <div class="text-sm font-medium text-foreground">{{ currentYear }}</div>
            <div class="flex items-center gap-1">
                <button class="p-1 hover:bg-muted rounded-md transition-colors" @click="prevMonth">
                    <ChevronLeftIcon class="w-4 h-4 text-muted-foreground" />
                </button>
                <div class="text-xs font-medium text-muted-foreground min-w-[32px] text-center select-none">{{
                    currentMonth }}</div>
                <button class="p-1 hover:bg-muted rounded-md transition-colors" :disabled="isFuture" @click="nextMonth">
                    <ChevronRightIcon class="w-4 h-4 text-muted-foreground" :class="{ 'opacity-30': isFuture }" />
                </button>
            </div>
        </div>

        <!-- Calendar with Week Numbers -->
        <div class="w-full">
            <!-- Headers -->
            <div class="grid grid-cols-8 gap-1.5 mb-1.5 align-middle">
                <!-- Empty corner cell -->
                <div class="w-full"></div>
                <!-- Weekday Headers -->
                <div
v-for="day in weekDays" :key="day"
                    class="text-[10px] text-muted-foreground text-center font-medium">
                    {{ day }}
                </div>
            </div>

            <!-- Weeks -->
            <div class="flex flex-col gap-1.5">
                <div
v-for="(week, wIndex) in calendarWeeks" :key="wIndex"
                    class="grid grid-cols-8 gap-1.5 items-center">
                    <!-- Week Number -->
                    <div class="text-[10px] text-muted-foreground text-center font-mono">
                        W{{ week.weekNum }}
                    </div>

                    <!-- Days -->
                    <div
v-for="(day, dIndex) in week.days" :key="dIndex"
                        class="aspect-square w-full flex items-center justify-center">
                        <div
v-if="day.isCurrentMonth" class="w-3 h-3 rounded-[2px] transition-all" :class="[
                            getCellClass(day.count),
                            { 'ring-1 ring-primary/50 ring-offset-1 ring-offset-background': isToday(day.date) },
                            day.count > 0 ? 'cursor-pointer hover:opacity-80' : 'cursor-default'
                        ]" :title="`${day.date}: ${day.count} 条记录`"
                            @click="day.count > 0 && $emit('day-click', day.date)">
                        </div>
                        <div v-else class="w-full h-full"></div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Legend -->
        <!-- Legend -->
        <div class="flex items-center justify-end gap-2 mt-4 pl-1">
            <span class="text-[10px] text-muted-foreground">{{ t('memo.less') }}</span>
            <div class="w-2.5 h-2.5 rounded-[2px] bg-secondary/80"></div>
            <div class="w-2.5 h-2.5 rounded-[2px] bg-primary/40"></div>
            <div class="w-2.5 h-2.5 rounded-[2px] bg-primary/60"></div>
            <div class="w-2.5 h-2.5 rounded-[2px] bg-primary/80"></div>
            <div class="w-2.5 h-2.5 rounded-[2px] bg-primary"></div>
            <span class="text-[10px] text-muted-foreground">{{ t('memo.more') }}</span>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import isoWeek from 'dayjs/plugin/isoWeek'
import weekOfYear from 'dayjs/plugin/weekOfYear'
import { ChevronLeftIcon, ChevronRightIcon } from '@heroicons/vue/24/outline'

dayjs.extend(isoWeek)
dayjs.extend(weekOfYear)
dayjs.locale('zh-cn')

interface Props {
    data: Record<string, number>
}

const props = defineProps<Props>()
const emit = defineEmits<{
    (e: 'day-click', date: string): void
}>()

const { t } = useI18n()

const currentDate = ref(dayjs())
const weekDays = ['一', '二', '三', '四', '五', '六', '日']

const currentYear = computed(() => currentDate.value.format('YYYY'))
const currentMonth = computed(() => currentDate.value.format('MM'))

const isFuture = computed(() => {
    return currentDate.value.add(1, 'month').startOf('month').isAfter(dayjs())
})

const calendarWeeks = computed(() => {
    const startOfMonth = currentDate.value.startOf('month')
    const endOfMonth = currentDate.value.endOf('month')

    // Start from the Monday of the first week
    let current = startOfMonth.startOf('isoWeek')
    // End at the Sunday of the last week
    const end = endOfMonth.endOf('isoWeek')

    const weeks = []

    while (current.isBefore(end) || current.isSame(end, 'day')) {
        const weekNum = current.isoWeek()
        const days = []

        for (let i = 0; i < 7; i++) {
            const dateStr = current.format('YYYY-MM-DD')
            days.push({
                date: dateStr,
                count: props.data[dateStr] || 0,
                isCurrentMonth: current.month() === startOfMonth.month()
            })
            current = current.add(1, 'day')
        }

        weeks.push({
            weekNum,
            days
        })
    }

    return weeks
})

function prevMonth() {
    currentDate.value = currentDate.value.subtract(1, 'month')
}

function nextMonth() {
    if (!isFuture.value) {
        currentDate.value = currentDate.value.add(1, 'month')
    }
}

function isToday(dateStr: string) {
    return dateStr === dayjs().format('YYYY-MM-DD')
}

function getCellClass(count: number): string {
    if (count === 0) return 'bg-secondary/80'
    if (count <= 1) return 'bg-primary/40'
    if (count <= 2) return 'bg-primary/50'
    if (count <= 4) return 'bg-primary/60'
    if (count <= 6) return 'bg-primary/70'
    if (count <= 8) return 'bg-primary/80'
    return 'bg-primary'
}
</script>
