import apiClient from './client';

// Ajustes del negocio (una sola fila por base de datos).
const configuracionApi = {
  obtener: () => apiClient.get('/configuracion'),
  actualizar: (nombreNegocio) => apiClient.put('/configuracion', { nombre_negocio: nombreNegocio }),
};

export default configuracionApi;
