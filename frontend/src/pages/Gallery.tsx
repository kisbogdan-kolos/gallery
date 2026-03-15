import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../api';
import { useAuth } from '../AuthContext';
import { Trash2, X } from 'lucide-react';

interface ImageRecord {
  id: number;
  name: string;
  uploaded: string;
  uploader: {
    username: string;
    id: number;
    displayname: string;
    admin: boolean;
  };
  image: string | null;
}

export default function Gallery() {
  const [images, setImages] = useState<ImageRecord[]>([]);
  const [error, setError] = useState('');
  const [selectedImage, setSelectedImage] = useState<ImageRecord | null>(null);
  const { user } = useAuth();

  const loadImages = async () => {
    try {
      const res = await fetchWithAuth('/api/image');
      if (res.ok) {
        const data = await res.json();
        setImages(data || []);
      } else {
        setError('Failed to load images');
      }
    } catch (e) {
      setError('Network error');
    }
  };

  useEffect(() => {
    loadImages();
  }, []);

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this image?')) return;
    try {
      const res = await fetchWithAuth(`/api/image/${id}`, { method: 'DELETE' });
      if (res.ok) {
        loadImages();
      } else {
        const d = await res.json();
        alert(d.error || 'Failed to delete');
      }
    } catch (e) {
      alert('Network error');
    }
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Gallery</h1>
      </div>
      {error && <div className="text-red-600 mb-4">{error}</div>}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
        {images.map((img) => (
          <div key={img.id} className="bg-white rounded-lg shadow overflow-hidden flex flex-col">
            <div className="aspect-w-1 aspect-h-1 w-full bg-gray-200">
              {img.image ? (
                <img
                  src={`/api/storage/${img.image}`}
                  alt={img.name}
                  onClick={() => setSelectedImage(img)}
                  className="w-full h-48 object-cover cursor-pointer transition-opacity hover:opacity-90"
                />
              ) : (
                <div className="w-full h-48 flex items-center justify-center text-gray-500">
                  No image data
                </div>
              )}
            </div>
            <div className="p-4 flex-1 flex flex-col justify-between">
              <div>
                <h3 className="text-lg font-medium text-gray-900 truncate" title={img.name}>
                  {img.name}
                </h3>
                <p className="text-sm text-gray-500 mt-1">Uploaded by: {img.uploader.displayname}</p>
                <p className="text-xs text-gray-400 mt-1">
                  {new Date(img.uploaded).toLocaleString()}
                </p>
              </div>
              {(user?.admin || user?.id === img.uploader.id) && (
                <div className="mt-4 flex justify-end">
                  <button
                    onClick={() => handleDelete(img.id)}
                    className="text-red-600 hover:text-red-900 p-1"
                    title="Delete image"
                  >
                    <Trash2 className="w-5 h-5" />
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}
        {images.length === 0 && !error && (
          <div className="col-span-full text-center text-gray-500 py-12">
            No images found.
          </div>
        )}
      </div>

      {selectedImage && selectedImage.image && (
        <div 
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-80"
          onClick={() => setSelectedImage(null)}
        >
          <button 
            className="absolute top-4 right-4 text-white hover:text-gray-300 p-2"
            onClick={() => setSelectedImage(null)}
          >
            <X className="w-8 h-8" />
          </button>
          <div className="flex flex-col items-center max-w-full max-h-full">
            <img
              src={`/api/storage/${selectedImage.image}`}
              alt={selectedImage.name}
              className="max-h-[85vh] max-w-full object-contain rounded"
              onClick={(e) => e.stopPropagation()}
            />
            <div className="text-white mt-4 text-center">
              <p className="text-xl font-medium">{selectedImage.name}</p>
              <p className="text-sm mt-1 text-gray-300">
                Uploaded by {selectedImage.uploader.displayname}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
