import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import { useAuthStore } from './store/authStore'
import AdminCourses from './pages/AdminCourses'
import AdminDashboard from './pages/AdminDashboard'
import AdminDecks from './pages/AdminDecks'
import AdminPodcasts from './pages/AdminPodcasts'
import AdminUsers from './pages/AdminUsers'
import CourseDeckPage from './pages/CourseDeckPage'
import Courses from './pages/Courses'
import Login from './pages/Login'
import Progress from './pages/Progress'
import Reader from './pages/Reader'
import ReaderTextPage from './pages/ReaderTextPage'
import Register from './pages/Register'
import VocabularyDeckPage from './pages/VocabularyDeckPage'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore()
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" />
}

function AdminRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, user } = useAuthStore()

  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }

  if (user?.role !== 'admin') {
    return <Navigate to="/" />
  }

  return <>{children}</>
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/"
          element={
            <PrivateRoute>
              <Layout />
            </PrivateRoute>
          }
        >
          <Route index element={<Navigate to="/courses" replace />} />
          <Route path="courses" element={<Courses />} />
          <Route path="courses/:id" element={<CourseDeckPage />} />
          <Route path="deck/:courseSlug" element={<CourseDeckPage />} />
          <Route path="deck/:courseSlug/:deckSlug" element={<CourseDeckPage />} />
          <Route path="vocabulary" element={<VocabularyDeckPage />} />
          <Route path="progress" element={<Progress />} />
          <Route path="reader" element={<Reader />} />
          <Route path="reader/:id" element={<ReaderTextPage />} />
          <Route
            path="admin"
            element={
              <AdminRoute>
                <AdminDashboard />
              </AdminRoute>
            }
          />
          <Route
            path="admin/courses"
            element={
              <AdminRoute>
                <AdminCourses />
              </AdminRoute>
            }
          />
          <Route
            path="admin/decks/:courseId"
            element={
              <AdminRoute>
                <AdminDecks />
              </AdminRoute>
            }
          />
          <Route
            path="admin/users"
            element={
              <AdminRoute>
                <AdminUsers />
              </AdminRoute>
            }
          />
          <Route
            path="admin/podcasts"
            element={
              <AdminRoute>
                <AdminPodcasts />
              </AdminRoute>
            }
          />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
