type TrainingModeCard = {
  is_custom: boolean
  image_url?: string | null
}

type TrainingModeProgress = {
  mode_view?: boolean
  mode_choice?: boolean
  mode_with_photo?: boolean
  mode_russian?: boolean
  mode_constructor?: boolean
} | null | undefined

export type ActiveTrainingMode = 'view' | 'choice' | 'with_photo' | 'russian' | 'constructor'
export type TrainingMode = ActiveTrainingMode | 'completed'

export const trainingModeMeta: Record<ActiveTrainingMode, { label: string; shortLabel: string }> = {
  view: {
    label: 'Просмотр',
    shortLabel: 'Просмотр',
  },
  choice: {
    label: 'Выбор 1 из 3',
    shortLabel: '1 из 3',
  },
  with_photo: {
    label: 'По картинке',
    shortLabel: 'Картинка',
  },
  russian: {
    label: 'Перевод на английский',
    shortLabel: 'Перевод',
  },
  constructor: {
    label: 'Конструктор',
    shortLabel: 'Конструктор',
  },
}

function hasImage(card: TrainingModeCard) {
  return Boolean(card.image_url && card.image_url.trim())
}

export function getTrainingModesForCard(card: TrainingModeCard): ActiveTrainingMode[] {
  if (card.is_custom) {
    return ['choice', 'constructor']
  }

  const modes: ActiveTrainingMode[] = ['view', 'choice']
  if (hasImage(card)) {
    modes.push('with_photo')
  }
  modes.push('russian', 'constructor')

  return modes
}

export function isTrainingModeCompleted(progress: TrainingModeProgress, mode: ActiveTrainingMode) {
  if (!progress) {
    return false
  }

  switch (mode) {
    case 'view':
      return Boolean(progress.mode_view)
    case 'choice':
      return Boolean(progress.mode_choice)
    case 'with_photo':
      return Boolean(progress.mode_with_photo)
    case 'russian':
      return Boolean(progress.mode_russian)
    case 'constructor':
      return Boolean(progress.mode_constructor)
    default:
      return false
  }
}

export function getTrainingProgressForCard(card: TrainingModeCard, progress: TrainingModeProgress) {
  const modes = getTrainingModesForCard(card)
  if (modes.length === 0) {
    return 100
  }

  const completedCount = modes.filter((mode) => isTrainingModeCompleted(progress, mode)).length
  return Math.round((completedCount / modes.length) * 100)
}

export function getCurrentTrainingModeForCard(card: TrainingModeCard, progress: TrainingModeProgress): TrainingMode {
  const nextMode = getTrainingModesForCard(card).find((mode) => !isTrainingModeCompleted(progress, mode))
  return nextMode || 'completed'
}

export function getTrainingModeStepLabel(card: TrainingModeCard, mode: TrainingMode | string) {
  if (mode === 'completed') {
    return 'Готово'
  }

  if (!(mode in trainingModeMeta)) {
    return mode
  }

  const modes = getTrainingModesForCard(card)
  const typedMode = mode as ActiveTrainingMode
  const modeIndex = modes.findIndex((currentMode) => currentMode === typedMode)
  if (modeIndex === -1) {
    return trainingModeMeta[typedMode].label
  }

  return `${modeIndex + 1}/${modes.length} ${trainingModeMeta[typedMode].label}`
}
