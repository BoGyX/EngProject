import { UserCourse } from '../services/studyService'

function getUserCourseTimestamp(userCourse: UserCourse) {
  return new Date(userCourse.last_opened_at || 0).getTime()
}

export function normalizeCourseProgress(value?: number) {
  return Math.max(0, Math.min(100, Math.round(Number(value || 0))))
}

export function sortUserCourses(userCourses: UserCourse[]) {
  return [...userCourses].sort((left, right) => {
    const activeDelta = Number(right.is_active) - Number(left.is_active)
    if (activeDelta !== 0) {
      return activeDelta
    }

    const timeDelta = getUserCourseTimestamp(right) - getUserCourseTimestamp(left)
    if (timeDelta !== 0) {
      return timeDelta
    }

    const attemptDelta = Number(right.attempt_number || 0) - Number(left.attempt_number || 0)
    if (attemptDelta !== 0) {
      return attemptDelta
    }

    return Number(right.id) - Number(left.id)
  })
}

export function buildLatestUserCourseMap(userCourses: UserCourse[]) {
  const latestCourseMap = new Map<number, UserCourse>()

  sortUserCourses(userCourses).forEach((userCourse) => {
    if (!latestCourseMap.has(userCourse.course_id)) {
      latestCourseMap.set(userCourse.course_id, userCourse)
    }
  })

  return latestCourseMap
}

export function getBlockingUserCourse(userCourses: UserCourse[]) {
  const latestCourses = Array.from(buildLatestUserCourseMap(userCourses).values())

  return (
    sortUserCourses(latestCourses).find((userCourse) => normalizeCourseProgress(userCourse.progress_percentage) < 100) || null
  )
}
