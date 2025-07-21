import apiClient from './client';

export const obtenerHistorial = (entidad_tipo) => {
  return apiClient.get(`historial/${entidad_tipo}/`);
};
