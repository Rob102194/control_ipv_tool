import apiClient from './client';

// API para gestionar el Inventario Diario (IPV).
const ipvApi = {
  // Obtiene el estado del inventario para una fecha específica.
  getEstado: (fecha) => {
    return apiClient.get(`/ipv/estado?fecha=${fecha}`);
  },

  // Solicita el cálculo del consumo basado en las ventas de una fecha.
  calcularConsumo: (fecha) => {
    return apiClient.get(`/ipv/calcular-consumo?fecha=${fecha}`);
  },

  // Guarda el estado completo del inventario para una fecha.
  guardar: (data) => {
    return apiClient.post('/ipv/guardar', data);
  },

  // Obtiene los modelos de IPV.
  getModelos: () => {
    return apiClient.get('/ipv/modelos');
  },

  // Guarda un modelo de IPV para un área.
  guardarModelo: (modelo) => {
    return apiClient.post('/ipv/modelos', modelo);
  },

  // Obtiene todos los registros de IPV.
  getRegistros: () => {
    return apiClient.get('/ipv/registros');
  },

  // Genera un reporte de IPV para una fecha.
  generarReporte: (fecha) => {
    return apiClient.get(`/ipv/reporte?fecha=${fecha}`);
  }
};

export default ipvApi;
