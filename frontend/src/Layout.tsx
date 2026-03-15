
import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext';
import { LogOut, Image as ImageIcon, Upload, User as UserIcon } from 'lucide-react';

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-100">
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Link to="/" className="flex-shrink-0 flex items-center gap-2">
                <ImageIcon className="w-8 h-8 text-blue-600" />
                <span className="font-bold text-xl text-gray-900">Gallery</span>
              </Link>
            </div>
            <div className="flex items-center gap-4">
              {user ? (
                <>
                  <Link
                    to="/upload"
                    className="inline-flex items-center gap-2 px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700"
                  >
                    <Upload className="w-4 h-4" />
                    Upload
                  </Link>
                  <div className="flex items-center gap-2 text-gray-700">
                    <UserIcon className="w-5 h-5" />
                    <span>{user.displayname}</span>
                  </div>
                  <button
                    onClick={handleLogout}
                    className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md"
                  >
                    <LogOut className="w-4 h-4" />
                    Logout
                  </button>
                </>
              ) : (
                <>
                  <Link to="/login" className="text-gray-600 hover:text-gray-900 font-medium">
                    Login
                  </Link>
                  <Link
                    to="/register"
                    className="inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                  >
                    Register
                  </Link>
                </>
              )}
            </div>
          </div>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  );
}
