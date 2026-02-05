<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import {
  CalendarRoot,
  type CalendarRootEmits,
  type CalendarRootProps,
  useForwardPropsEmits,
} from 'radix-vue'
import { getLocalTimeZone } from '@internationalized/date'
import dayjs from 'dayjs'
import weekOfYear from 'dayjs/plugin/weekOfYear'
import { cn } from '@/lib/utils'
import CalendarHeader from './CalendarHeader.vue'
import CalendarGrid from './CalendarGrid.vue'
import CalendarGridHead from './CalendarGridHead.vue'
import CalendarGridBody from './CalendarGridBody.vue'
import CalendarGridRow from './CalendarGridRow.vue'
import CalendarHeadCell from './CalendarHeadCell.vue'
import CalendarCell from './CalendarCell.vue'
import CalendarCellTrigger from './CalendarCellTrigger.vue'
import CalendarPrevButton from './CalendarPrevButton.vue'
import CalendarNextButton from './CalendarNextButton.vue'
import CalendarHeading from './CalendarHeading.vue'

dayjs.extend(weekOfYear)

const props = withDefaults(defineProps<CalendarRootProps & { class?: HTMLAttributes['class']; showWeekNumber?: boolean }>(), {
  weekStartsOn: 1,
})

const emits = defineEmits<CalendarRootEmits>()

const delegatedProps = computed(() => {
  const { class: _, showWeekNumber: __, ...delegated } = props

  return delegated
})

const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <CalendarRoot
    v-slot="{ grid, weekDays }"
    :class="cn('p-3', props.class)"
    v-bind="forwarded"
  >
    <CalendarHeader>
      <CalendarPrevButton />
      <CalendarHeading />
      <CalendarNextButton />
    </CalendarHeader>

    <div class="flex flex-col gap-y-4 mt-4 sm:flex-row sm:gap-x-4 sm:gap-y-0">
      <CalendarGrid v-for="month in grid" :key="month.value.toString()">
        <CalendarGridHead>
          <CalendarGridRow>
             <CalendarHeadCell v-if="props.showWeekNumber" class="w-10 pb-2 text-sm font-normal text-muted-foreground/60">
              <span class="text-[0.7rem]"></span>
            </CalendarHeadCell>
            <CalendarHeadCell
              v-for="day in weekDays"
              :key="day"
              class="pb-2 text-sm font-normal text-muted-foreground/60"
            >
              <span class="text-[0.7rem]">{{ day }}</span>
            </CalendarHeadCell>
          </CalendarGridRow>
        </CalendarGridHead>
        <CalendarGridBody>
          <CalendarGridRow v-for="(weekDates, index) in month.rows" :key="`weekDate-${index}`">
            <CalendarCell v-if="props.showWeekNumber" :date="weekDates[0]" class="w-10 p-0 text-center text-[0.8rem] text-muted-foreground/50 flex items-center justify-center">
              {{ dayjs(weekDates[0].toDate(getLocalTimeZone())).week() }}
            </CalendarCell>
            <CalendarCell
              v-for="weekDate in weekDates"
              :key="weekDate.toString()"
              :date="weekDate"
            >
              <CalendarCellTrigger
                :day="weekDate"
                :month="month.value"
              />
            </CalendarCell>
          </CalendarGridRow>
        </CalendarGridBody>
      </CalendarGrid>
    </div>
  </CalendarRoot>
</template>
<!-- Updated week start logic -->