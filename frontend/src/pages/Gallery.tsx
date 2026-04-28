import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { fetchWithAuth } from '../api';
import { useAuth } from '../AuthContext';
import { Trash2, X, ExternalLink, ScanText } from 'lucide-react';

interface TextRecord {
  text: string;
  xmin: number;
  ymin: number;
  xmax: number;
  ymax: number;
}

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
  text?: TextRecord[];
  ocrtime: string | null;
}

export default function Gallery() {
  const [images, setImages] = useState<ImageRecord[]>([]);
  const [error, setError] = useState('');
  const [imageDimensions, setImageDimensions] = useState<{width: number, height: number} | null>(null);
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const selectedImage = images.find((img) => img.id.toString() === id) || null;

  useEffect(() => {
    setImageDimensions(null);
  }, [selectedImage]);

  const closeImage = () => {
    navigate('/');
  };

  const openImage = (img: ImageRecord) => {
    navigate(`/image/${img.id}`);
  };

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

  const handleOcr = async (id: number) => {
    if (!confirm('Are you sure you want to run OCR on this image?')) return;
    try {
      const res = await fetchWithAuth(`/api/image/${id}/ocr`, { method: 'POST' });
      if (res.ok) {
        loadImages();
      } else {
        const d = await res.json();
        alert(d.error || 'Failed to start OCR');
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
                  onClick={() => openImage(img)}
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
              <div className="mt-4 flex justify-between items-center">
                {img.image && (
                  <a
                    href={`/api/storage/${img.image}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-indigo-600 hover:text-indigo-900 p-1 flex items-center"
                    title="Open in new tab"
                  >
                    <ExternalLink className="w-5 h-5 mr-1" />
                  </a>
                )}
                {(user?.admin || user?.id === img.uploader.id) && (
                  <div className="ml-auto flex items-center space-x-1">
                    <button
                      onClick={() => handleOcr(img.id)}
                      className="text-blue-600 hover:text-blue-900 p-1"
                      title="Run OCR"
                    >
                      <ScanText className="w-5 h-5" />
                    </button>
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
          onClick={closeImage}
        >
          <button 
            className="absolute top-4 right-4 text-white hover:text-gray-300 p-2"
            onClick={closeImage}
          >
            <X className="w-8 h-8" />
          </button>
          <div className="flex flex-col items-center max-w-full max-h-full">
            <div className="relative inline-block max-h-[85vh] max-w-full">
              <img
                src={`/api/storage/${selectedImage.image}`}
                alt={selectedImage.name}
                className="max-h-[85vh] max-w-full block rounded"
                onClick={(e) => e.stopPropagation()}
                onLoad={(e) => {
                  setImageDimensions({
                    width: e.currentTarget.naturalWidth,
                    height: e.currentTarget.naturalHeight
                  });
                }}
              />
              {imageDimensions && selectedImage.text && selectedImage.text.length > 0 && (
                <svg
                  className="absolute inset-0 w-full h-full"
                  viewBox={`0 0 ${imageDimensions.width} ${imageDimensions.height}`}
                  onClick={(e) => e.stopPropagation()}
                >
                  {selectedImage.text.map((t, i) => (
                    <g key={i}>
                      <rect
                        x={t.xmin}
                        y={t.ymin}
                        width={t.xmax - t.xmin}
                        height={t.ymax - t.ymin}
                        fill="rgba(0,0,0,0.2)"
                        stroke="rgb(239, 68, 68)"
                        strokeWidth="2"
                      />
                      <text
                        x={t.xmin + (t.xmax - t.xmin) / 2}
                        y={t.ymin + (t.ymax - t.ymin) / 2}
                        fill="rgba(255,255,255,0.8)"
                        fontSize={(t.ymax - t.ymin) * 0.8}
                        dominantBaseline="central"
                        textAnchor="middle"
                        textLength={t.xmax - t.xmin}
                        lengthAdjust="spacingAndGlyphs"
                        className="select-text cursor-text"
                      >
                        {t.text}
                      </text>
                    </g>
                  ))}
                </svg>
              )}
            </div>
            <div className="text-white mt-4 text-center">
              <p className="text-xl font-medium">{selectedImage.name}</p>
              <p className="text-sm mt-1 text-gray-300">
                Uploaded by {selectedImage.uploader.displayname} at {new Date(selectedImage.uploaded).toLocaleString()}
              </p>
              <p className="text-xs mt-1 text-gray-400">
                OCR finished at {selectedImage.ocrtime ? new Date(selectedImage.ocrtime).toLocaleString() : 'OCR not run yet'}
              </p>
              <a
                href={`/api/storage/${selectedImage.image}`}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center mt-3 px-3 py-1 bg-white bg-opacity-20 hover:bg-opacity-30 rounded-md text-sm text-white transition-colors"
                onClick={(e) => e.stopPropagation()}
                title="Open raw image"
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                Open original
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
