import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { config } from '../config'

const adminCards = [
  {
    title: 'Управление курсами',
    description: 'Создание, редактирование и публикация курсов с быстрым переходом к декам.',
    href: '/admin/courses',
    badge: 'Контент',
  },
  {
    title: 'Подкасты',
    description: 'Все mp3 из reader, привязанные к курсам, с фильтром по курсу и прослушиванием прямо в панели.',
    href: '/admin/podcasts',
    badge: 'Аудио',
  },
  {
    title: 'Пользователи',
    description: 'Просмотр зарегистрированных пользователей и контроль ролей внутри платформы.',
    href: '/admin/users',
    badge: 'Доступ',
  },
]

export default function AdminDashboard() {
  const { user } = useAuthStore()

  if (user?.role !== 'admin') {
    return <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-red-700">У вас нет доступа к админ панели</div>
  }

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 shadow-xl">
        <div className="grid gap-8 p-6 lg:grid-cols-[minmax(0,1fr)_280px] lg:p-8">
          <div className="space-y-5">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Admin
              </span>
              <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">Внутренняя панель</span>
            </div>

            <div>
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">Админ панель</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Быстрый вход в разделы управления курсами, подкастами и пользователями. Панель собрана в том же визуальном
                стиле, что и рабочие страницы продукта.
              </p>
            </div>
          </div>

          <div className="rounded-[24px] border border-rose-100 bg-white/80 p-6 shadow-lg">
            <p className="text-xs uppercase tracking-[0.18em] text-rose-500">Текущий администратор</p>
            <p className="mt-3 text-2xl font-bold text-text-light">{user?.name || user?.email}</p>
            <p className="mt-3 text-sm text-slate-600">Используйте карточки ниже для перехода к управлению содержимым и пользователями.</p>
          </div>
        </div>
      </section>

      <section className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
        {adminCards.map((card) => (
          <Link
            key={card.href}
            to={card.href}
            className="group rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md transition-all hover:-translate-y-1 hover:border-rose-200 hover:shadow-xl"
          >
            <span className="rounded-full bg-rose-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-rose-700">
              {card.badge}
            </span>
            <h2 className="mt-4 text-2xl font-semibold text-text-light">{card.title}</h2>
            <p className="mt-3 text-sm leading-7 text-slate-600">{card.description}</p>
            <span className="mt-5 inline-flex text-sm font-semibold text-link-light transition-colors group-hover:text-link-dark">
              Открыть раздел
            </span>
          </Link>
        ))}
      </section>

      <section className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
        <h2 className="text-xl font-semibold text-text-light">Быстрые действия</h2>
        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          <Link
            to="/admin/courses"
            className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50"
          >
            Создать или отредактировать курс
          </Link>
          <Link
            to="/admin/podcasts"
            className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50"
          >
            Открыть подкасты
          </Link>
          <a
            href={`${config.baseUrl}/swagger/index.html`}
            target="_blank"
            rel="noopener noreferrer"
            className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50"
          >
            Открыть Swagger документацию
          </a>
        </div>
      </section>
    </div>
  )
}
