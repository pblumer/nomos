import { FormEvent, useState } from "react";

type LoginProps = {
  apiBaseUrl: string;
  onLoggedIn: (token: string, username: string) => void;
};

type LoginResponse = {
  token: string;
  username: string;
  expires_at: number;
};

function Login({ apiBaseUrl, onLoggedIn }: LoginProps) {
  const [username, setUsername] = useState("nomos");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: username.trim(), password }),
      });
      if (!response.ok) {
        let detail = "Login failed";
        try {
          const payload = (await response.json()) as { detail?: string };
          if (payload?.detail) detail = payload.detail;
        } catch {
          /* ignore */
        }
        setError(detail);
        return;
      }
      const data = (await response.json()) as LoginResponse;
      onLoggedIn(data.token, data.username);
    } catch {
      setError("Request failed");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 font-sans">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm bg-white rounded-xl shadow-xl border border-slate-200 px-8 py-10"
        aria-label="Sign in"
      >
        <div className="flex items-center gap-3 mb-6">
          <div className="w-9 h-9 bg-indigo-500 rounded-lg flex items-center justify-center">
            <span aria-hidden="true" className="material-symbols-outlined text-white text-sm">account_balance</span>
          </div>
          <h1 className="text-xl font-semibold text-slate-900">Nomos</h1>
        </div>
        <p className="text-sm text-slate-600 mb-6">Please sign in to continue.</p>

        <label className="block text-xs font-medium text-slate-600 mb-1" htmlFor="login-username">Username</label>
        <input
          id="login-username"
          type="text"
          autoComplete="username"
          value={username}
          onChange={(event) => setUsername(event.target.value)}
          className="w-full mb-4 rounded-md border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          required
        />

        <label className="block text-xs font-medium text-slate-600 mb-1" htmlFor="login-password">Password</label>
        <input
          id="login-password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="w-full mb-4 rounded-md border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          required
        />

        {error ? (
          <div role="alert" className="mb-4 text-sm text-red-700 bg-red-50 border border-red-100 rounded-md px-3 py-2">
            {error}
          </div>
        ) : null}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-md bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white text-sm font-medium px-4 py-2"
        >
          {submitting ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}

export default Login;
