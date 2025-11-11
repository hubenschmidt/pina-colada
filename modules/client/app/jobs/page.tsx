"use client"

import { useState, useEffect } from 'react'
import JobTracker from '../../components/JobTracker/JobTracker'
import LoginForm from '../../components/JobTracker/LoginForm'
import { isAuthenticated, logout } from '../../lib/auth'
import { LogOut } from 'lucide-react'

const JobsPage = () => {
  const [authenticated, setAuthenticated] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setAuthenticated(isAuthenticated())
    setLoading(false)
  }, [])

  const handleLoginSuccess = () => {
    setAuthenticated(true)
  }

  const handleLogout = () => {
    logout()
    setAuthenticated(false)
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-blue-50 flex items-center justify-center">
        <p className="text-zinc-600">Loading...</p>
      </div>
    )
  }

  if (!authenticated) {
    return <LoginForm onLoginSuccess={handleLoginSuccess} />
  }

  return (
    <main className="min-h-screen bg-blue-50 py-8 px-4">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-end mb-4">
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-zinc-300 rounded-lg text-zinc-700 hover:bg-zinc-50 transition-colors"
          >
            <LogOut size={18} />
            Logout
          </button>
        </div>
        <JobTracker />
      </div>
    </main>
  )
}

export default JobsPage
