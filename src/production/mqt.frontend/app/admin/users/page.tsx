"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { adminService } from "@/services/api/adminService";
import { Loader2, AlertCircle, X, Trash2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { User } from "@/types/admin";

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [showUserForm, setShowUserForm] = useState(false);
  const [showDeleteUserConfirm, setShowDeleteUserConfirm] = useState<string | null>(null);
  const [createUserSuccess, setCreateUserSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [userFormData, setUserFormData] = useState({
    username: "",
    email: "",
    password: "",
    role: "user" as "admin" | "user",
  });

  const loadUsers = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await adminService.getAllUsers();
      setUsers(data?.users || []);
    } catch (err) {
      console.error("Error loading users:", err);
      setError(err instanceof Error ? err.message : "Failed to load users");
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  const handleCreateUser = async () => {
    const emailUsed = userFormData.email;
    try {
      setLoading(true);
      setError(null);
      setCreateUserSuccess(null);
      const resp = await adminService.registerAdmin(userFormData);
      setShowUserForm(false);
      setUserFormData({ username: "", email: "", password: "", role: "user" });
      await loadUsers();
      if (resp.requires_verification) {
        const email = resp.email ?? emailUsed;
        setCreateUserSuccess(`User created. An email verification link/code was sent to ${email}. They must verify before logging in.`);
      } else {
        setCreateUserSuccess("User created successfully.");
      }
      setTimeout(() => setCreateUserSuccess(null), 8000);
    } catch (err) {
      console.error("Error creating user:", err);
      setError(err instanceof Error ? err.message : "Failed to create user");
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateUserRole = async (userId: string, role: "admin" | "user") => {
    try {
      setLoading(true);
      setError(null);
      await adminService.updateUserRole(userId, role);
      await loadUsers();
    } catch (err) {
      console.error("Error updating user role:", err);
      setError(err instanceof Error ? err.message : "Failed to update user role");
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteUser = async (userId: string) => {
    try {
      setLoading(true);
      setError(null);
      await adminService.deleteUser(userId);
      setShowDeleteUserConfirm(null);
      await loadUsers();
    } catch (err) {
      console.error("Error deleting user:", err);
      setError(err instanceof Error ? err.message : "Failed to delete user");
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setError(null);
    await loadUsers();
  };

  return (
    <main className="pt-16 sm:pt-24 px-4 sm:px-6 lg:px-8 pb-12 sm:pb-16">
      <div className="max-w-7xl mx-auto">
        <div className="mb-6 sm:mb-8 flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-4xl font-light tracking-tight mb-2">User Management</h1>
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
            <Button onClick={() => setShowUserForm(true)}>Create User</Button>
          </div>
        </div>

        {error && (
          <div className="mb-6 p-4 border border-red-500/20 bg-red-500/10 rounded-lg flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-400" />
            <p className="text-sm text-red-400 font-light">{error}</p>
          </div>
        )}

        {createUserSuccess && (
          <div className="mb-6 p-4 border border-green-500/30 rounded-lg bg-green-500/10 text-green-400 text-sm font-light flex items-center justify-between">
            <span>{createUserSuccess}</span>
            <button onClick={() => setCreateUserSuccess(null)} className="text-green-400/80 hover:text-green-400">
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        {showUserForm && (
          <div className="border border-white/10 rounded-lg p-6 bg-black/50 mb-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-light">Create New User</h3>
              <button onClick={() => setShowUserForm(false)} className="text-white/60 hover:text-white">
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input
                placeholder="Username"
                value={userFormData.username}
                onChange={(e) => setUserFormData({ ...userFormData, username: e.target.value })}
                className="bg-black/50 border-white/10"
              />
              <Input
                type="email"
                placeholder="Email"
                value={userFormData.email}
                onChange={(e) => setUserFormData({ ...userFormData, email: e.target.value })}
                className="bg-black/50 border-white/10"
              />
              <Input
                type="password"
                placeholder="Password"
                value={userFormData.password}
                onChange={(e) => setUserFormData({ ...userFormData, password: e.target.value })}
                className="bg-black/50 border-white/10"
              />
              <select
                value={userFormData.role}
                onChange={(e) => setUserFormData({ ...userFormData, role: e.target.value as "admin" | "user" })}
                className="flex h-9 w-full rounded-md border border-white bg-black px-3 py-1 text-sm"
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <div className="flex gap-2 mt-4">
              <Button onClick={handleCreateUser}>Create</Button>
              <Button variant="outline" onClick={() => setShowUserForm(false)}>Cancel</Button>
            </div>
          </div>
        )}

        {loading && !users.length ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 text-white/60 animate-spin" />
          </div>
        ) : (
          <div className="border border-white/10 rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-black/50 border-b border-white/10">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-light">Username</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Email</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Role</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Active</th>
                  <th className="px-4 py-3 text-left text-sm font-light">Actions</th>
                </tr>
              </thead>
              <tbody>
                {(users || []).map((u) => (
                  <tr key={u.user_id} className="border-b border-white/10 hover:bg-white/5">
                    <td className="px-4 py-3 text-sm font-light">
                      <Link href={`/admin/users/${u.user_id}`} className="text-orange-400 hover:text-orange-300 hover:underline">
                        {u.username}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm font-light">{u.email}</td>
                    <td className="px-4 py-3 text-sm font-light">
                      <select
                        value={u.role}
                        onChange={(e) => handleUpdateUserRole(u.user_id, e.target.value as "admin" | "user")}
                        className="bg-black border border-white rounded px-2 py-1 text-sm"
                      >
                        <option value="user">User</option>
                        <option value="admin">Admin</option>
                      </select>
                    </td>
                    <td className="px-4 py-3 text-sm font-light">{u.active ? "Yes" : "No"}</td>
                    <td className="px-4 py-3 text-sm font-light">
                      <button
                        onClick={() => setShowDeleteUserConfirm(u.user_id)}
                        className="text-red-400 hover:text-red-300 transition-colors"
                        title="Delete user"
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

        {showDeleteUserConfirm && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-black border border-white/10 rounded-lg p-6 max-w-md w-full mx-4">
              <h3 className="text-lg font-light mb-4">Confirm Delete</h3>
              <p className="text-white/60 font-light mb-6">
                Are you sure you want to delete this user? This action cannot be undone.
              </p>
              <div className="flex gap-2">
                <Button variant="destructive" onClick={() => handleDeleteUser(showDeleteUserConfirm)}>
                  Delete
                </Button>
                <Button variant="outline" onClick={() => setShowDeleteUserConfirm(null)}>
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
