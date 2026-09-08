import axios from 'axios';

// URL base de la API. Por defecto es relativa ('/api'): funciona igual dentro
// del escritorio (Wails sirve el SPA y la API en el mismo origen) y en un
// despliegue web. En `npm run dev` el proxy de Vite (vite.config.js) redirige
// '/api' al servidor Go local. VITE_API_BASE_URL permite apuntar a otro host.
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Interceptor para manejar errores de respuesta de forma centralizada.
apiClient.interceptors.response.use(
  response => response,
  error => {
    // Aquí se podrían manejar errores específicos, como 401, 403, 500, etc.
    // Por ejemplo, redirigir al login si se recibe un 401 Unauthorized.
    console.error('Error en la llamada a la API:', error.response?.data || error.message);
    
    // Rechaza la promesa para que el error pueda ser capturado por el llamador.
    return Promise.reject(error);
  }
);

export default apiClient;
