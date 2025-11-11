"use client";

import { useState, FormEvent } from "react";
import { login } from "../../lib/auth";
import { Lock } from "lucide-react";

type LoginFormProps = {
  onLoginSuccess: () => void;
};

const LoginForm = ({ onLoginSuccess }: LoginFormProps) => {
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const success = await login(password);

      if (success) {
        onLoginSuccess();
      } else {
        setError("Invalid password. Please try again.");
        setPassword("");
      }
    } catch (error) {
      console.error("Login error:", error);
      setError("Login failed. Please try again.");
      setPassword("");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-blue-50 flex items-center justify-center px-4">
      <div className="max-w-md w-full">
        <div className="bg-white rounded-lg shadow-lg p-8">
          <div className="flex items-center justify-center mb-6">
            <div className="bg-gradient-to-r from-lime-500 to-yellow-400 p-4 rounded-full">
              <Lock size={32} className="text-blue-900" />
            </div>
          </div>

          <h1 className="text-2xl font-bold text-center text-blue-900 mb-2">
            Client Login
          </h1>
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label
                htmlFor="password"
                className="block text-sm font-medium text-zinc-700 mb-2"
              >
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password"
                className="w-full px-4 py-3 border border-zinc-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-lime-500"
                autoFocus
                required
              />
            </div>

            {error && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-800">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={isSubmitting || !password}
              className="w-full py-3 bg-gradient-to-r from-lime-500 to-yellow-400 text-blue-900 font-semibold rounded-lg hover:brightness-95 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              {isSubmitting ? "Logging in..." : "Login"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default LoginForm
