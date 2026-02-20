"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { adminService } from "@/services/api/adminService";
import { Loader2, AlertCircle, X, Trash2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Pi, User } from "@/types/admin";

export default function AdminPisPage() {
  const [pis, setPis] = useState<Pi[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [showPiForm, setShowPiForm] = useState(false);
  const [showDeletePiConfirm, setShowDeletePiConfirm] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [piFormData, setPiFormData] = useState({ pi_id: "", user_id: "" });

  const loadPis = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await adminService.getAllPis();
      setPis(data?.items || []);
    } catch (err) {
      console.error("Error loading PIs:", err);
      setError(err instanceof Error ? err.message : "Failed to load PIs");
      setPis([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadPis();
  }, [loadPis]);

  useEffect(() => {
    adminService.getAllUsers().then((d) => setUsers(d?.users || [])).catch(() => setUsers([]));
  }, []);

  const handleCreatePi = async () => {
    try {
      setLoading(true);
      setError(null);
      if (users.length === 0) {
        const usersData = await adminService.getAllUsers();
        setUsers(usersData.users);
      }
      await adminService.createPi({
        pi_id: piFormData.pi_id,
        user_id: piFormData.user_id || undefined,
      });
      setShowPiForm(false);
      setPiFormData({ pi_id: "", user_id: "" });
      await loadPis();
    } catch (err) {
      console.error("Error creating PI:", err);
      setError(err instanceof Error ? err.message : "Failed to create PI");
    } finally {
      setLoading(false);
    }
  };

  const handleUpdatePi = async (piId: string, userId: string) => {
    try {
      setLoading(true);
      setError(null);
      if (users.length === 0) {
        const usersData = await adminService.getAllUsers();
        setUsers(usersData.users);
      }
      await adminService.updatePi(piId, { user_id: userId || undefined });
      await loadPis();
    } catch (err) {
      console.error("Error updating PI:", err);
      setError(err instanceof Error ? err.message : "Failed to update PI");
    } finally {
      setLoading(false);
    }
  };

  const handleDeletePi = async (piId: string) => {
    try {
      setLoading(true);
      setError(null);
      await adminService.deletePi(piId, true);
      setShowDeletePiConfirm(null);
      await loadPis();
    } catch (err) {
      console.error("Error deleting PI:", err);
      setError(err instanceof Error ? err.message : "Failed to delete PI");
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setError(null);
    await loadPis();
  };

  return (
    <main className="pt-24 px-4 sm:px-6 lg:px-8 pb-16">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-light tracking-tight mb-2">PI Management</h1>
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleRefresh}
              disabled={loading}
              className="p-2 rounded-lg border border-white/20 hover:bg-white/5 text-white/70 hover:text-white disabled:opacity-50 transition-colors"
              title="Refresh data"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            </button>
            <Button onClick={() => setShowPiForm(true)}>Create PI</Button>
          </div>
        </div>

        {error && (
          <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-400" />
            <p className="text-sm text-red-400 font-light">{error}</p>
          </div>
        )}

        {showPiForm && (
          <div className="border border-white/10 rounded-lg p-6 bg-black/50 mb-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-light">Create New PI</h3>
              <button onClick={() => setShowPiForm(false)} className="text-white/60 hover:text-white">
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input
                placeholder="PI ID"
                value={piFormData.pi_id}
                onChange={(e) => setPiFormData({ ...piFormData, pi_id: e.target.value })}
                className="bg-black/50 border-white/10"
              />
              <select
                value={piFormData.user_id}
                onChange={(e) => setPiFormData({ ...piFormData, user_id: e.target.value })}
                className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
              >
                <option value="">No User (Unassigned)</option>
                {(users || []).map((u) => (
                  <option key={u.user_id} value={u.user_id}>
                    {u.username} ({u.email})
                  </option>
                ))}
              </select>
            </div>
            <div className="flex gap-2 mt-4">
              <Button onClick={handleCreatePi}>Create</Button>
              <Button variant="outline" onClick={() => setShowPiForm(false)}>Cancel</Button>
            </div>
          </div>
        )}

        {loading && !pis.length ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          </div>
        ) : (
          <div className="border border-white/10 rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-black/50 border-b border-white/10">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-light">PI ID</th>
                  <th className="px-4 py-3 text-left text-sm font-light">User</th>
                  <th className="px-4 py-3 text-left text-sm font-light">View</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Actions</th>
                </tr>
              </thead>
              <tbody>
                {(pis || []).map((pi) => (
                  <tr key={pi.pi_id} className="border-b border-white/10 hover:bg-white/5">
                    <td className="px-4 py-3 text-sm font-light">{pi.pi_id}</td>
                    <td className="px-4 py-3 text-sm font-light">
                      <select
                        value={String(pi.user_id ?? "")}
                        onChange={(e) => handleUpdatePi(pi.pi_id, e.target.value)}
                        className="bg-black border border-white rounded px-2 py-1 text-sm"
                      >
                        <option value="">Unassigned</option>
                        {(users || []).map((u) => (
                          <option key={u.user_id} value={String(u.user_id ?? "")}>
                            {u.username}
                          </option>
                        ))}
                      </select>
                    </td>
                    <td className="px-4 py-3 text-sm font-light">
                      <Link href={`/admin/devices?pi=${encodeURIComponent(pi.pi_id)}`}>
                        <Button size="sm" variant="outline" className="text-black bg-white/90 border-2 border-white hover:bg-white/20 hover:border-white/30 text-xs font-light h-8 px-4 flex items-center gap-1.5">
                          View
                        </Button>
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm font-light">
                      <button
                        onClick={() => setShowDeletePiConfirm(pi.pi_id)}
                        className="text-red-400 hover:text-red-300 transition-colors"
                        title="Delete PI"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {showDeletePiConfirm && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-black border border-white/10 rounded-lg p-6 max-w-md w-full mx-4">
              <h3 className="text-lg font-light mb-4">Confirm Delete</h3>
              <p className="text-white/60 font-light mb-6">
                Are you sure you want to delete this PI? This will also delete all associated devices and readings (cascade delete).
              </p>
              <div className="flex gap-2">
                <Button variant="destructive" onClick={() => handleDeletePi(showDeletePiConfirm)}>
                  Delete
                </Button>
                <Button variant="outline" onClick={() => setShowDeletePiConfirm(null)}>
                  Cancel
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
