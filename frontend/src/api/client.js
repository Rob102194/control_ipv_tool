import axios from 'axios';

// Obtiene la URL base de la API desde las variables de entorno de Vite.
// Si no está definida, utiliza un valor predeterminado para el desarrollo local.
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:5000/api';

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
