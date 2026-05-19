import { useState, useEffect } from 'react';
import { useAuth } from '../AuthContext';
import { fetchWithAuth } from '../api';
import type { User as UserType } from '../AuthContext';

export default function User() {
  const { user, checkAuth } = useAuth();
  const [editedDisplayname, setEditedDisplayname] = useState('');
  const [editedEmail, setEditedEmail] = useState('');
  const [editedPassword, setEditedPassword] = useState('');
  const [allUsers, setAllUsers] = useState<UserType[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (user) {
      setEditedDisplayname(user.displayname);
      setEditedEmail(user.email);
    }
    if (user?.admin) {
      fetchAllUsers();
    }
  }, [user]);

  const fetchAllUsers = async () => {
    try {
      const res = await fetchWithAuth('/api/user/all');
      if (res.ok) {
        const data = await res.json();
        setAllUsers(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch users:', err);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
      const updateData: Record<string, string> = {};
      if (editedDisplayname !== user?.displayname) {
        updateData.displayname = editedDisplayname;
      }
      if (editedEmail !== user?.email) {
        updateData.email = editedEmail;
      }
      if (editedPassword) {
        updateData.password = editedPassword;
      }

      if (Object.keys(updateData).length === 0) {
        setError('No changes to save');
        setLoading(false);
        return;
      }

      const res = await fetchWithAuth('/api/user/me', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updateData),
      });

      if (res.ok) {
        setEditedPassword('');
        setSuccess('Profile updated successfully');
        await checkAuth();
        if (user?.admin) {
          fetchAllUsers();
        }
      } else {
        const data = await res.json();
        setError(data.error || 'Failed to update profile');
      }
    } catch (err) {
      setError('Network error');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return <div className="text-center py-12 text-gray-500">Please log in to view your profile.</div>;
  }

  return (
    <div className="space-y-8">
      {/* Profile Edit Form */}
      <div className="bg-white p-8 rounded-lg shadow">
        <h2 className="text-2xl font-bold mb-6">My Profile</h2>

        {error && <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>}
        {success && <div className="bg-green-50 text-green-600 p-3 rounded mb-4">{success}</div>}

        <form onSubmit={handleSave} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Username - Read only */}
            <div>
              <label className="block text-sm font-medium text-gray-700">Username</label>
              <input
                type="text"
                disabled
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100 text-gray-500"
                value={user.username}
              />
            </div>

            {/* Display Name */}
            <div>
              <label className="block text-sm font-medium text-gray-700">Display Name</label>
              <input
                type="text"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                value={editedDisplayname}
                onChange={(e) => setEditedDisplayname(e.target.value)}
              />
            </div>

            {/* Email */}
            <div>
              <label className="block text-sm font-medium text-gray-700">Email</label>
              <input
                type="email"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                value={editedEmail}
                onChange={(e) => setEditedEmail(e.target.value)}
              />
            </div>

            {/* Registered Date - Read only */}
            <div>
              <label className="block text-sm font-medium text-gray-700">Registered</label>
              <input
                type="text"
                disabled
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100 text-gray-500"
                value={new Date(user.registered).toLocaleString()}
              />
            </div>
          </div>

          {/* Password */}
          <div>
            <label className="block text-sm font-medium text-gray-700">New Password (leave empty to keep current)</label>
            <input
              type="password"
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              value={editedPassword}
              onChange={(e) => setEditedPassword(e.target.value)}
              placeholder="Enter new password or leave empty"
            />
          </div>

          {/* Admin Badge */}
          {user.admin && (
            <div className="bg-blue-50 border border-blue-200 rounded p-3">
              <span className="text-sm font-semibold text-blue-700">Administrator</span>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            {loading ? 'Saving...' : 'Save Changes'}
          </button>
        </form>
      </div>

      {/* All Users List (Admin only) */}
      {user.admin && (
        <div className="bg-white p-8 rounded-lg shadow">
          <h2 className="text-2xl font-bold mb-6">All Users</h2>
          {allUsers.length === 0 ? (
            <p className="text-gray-500">No users found</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Username</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Display Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Registered</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Role</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {allUsers.map((u) => (
                    <tr key={u.id} className={u.id === user.id ? 'bg-blue-50' : ''}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{u.username}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">{u.displayname}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">{u.email || '-'}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">{new Date(u.registered).toLocaleString()}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        {u.admin ? (
                          <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">Admin</span>
                        ) : (
                          <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">User</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
