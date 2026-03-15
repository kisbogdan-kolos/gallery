
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import type { ReactNode } from 'react';
import { AuthProvider, useAuth } from './AuthContext';
import Layout from './Layout';
import Gallery from './pages/Gallery';
import Login from './pages/Login';
import Register from './pages/Register';
import Upload from './pages/Upload';

// A simple component to conditionally render based on auth state
function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  
  if (loading) return <div className="text-center py-12">Loading...</div>;
  if (!user) return <div className="text-center py-12 text-gray-500">Please log in to access this page.</div>;
  
  return children;
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Gallery />} />
            <Route path="login" element={<Login />} />
            <Route path="register" element={<Register />} />
            <Route path="upload" element={
              <RequireAuth>
                <Upload />
              </RequireAuth>
            } />
          </Route>
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
