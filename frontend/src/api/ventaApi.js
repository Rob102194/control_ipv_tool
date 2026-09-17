import client from './client';
import { fileToBase64 } from './base64';

export const getVentas = () => client.get('ventas/');
export const updateVenta = (id, data) => client.put(`ventas/${id}/`, data);
export const deleteVenta = (id) => client.delete(`ventas/${id}/`);
export const deleteVentas = (ids) => client.post('ventas/delete-multiple/', { ids });
export const importVentas = async (file, fecha) => {
    const archivo_base64 = await fileToBase64(file);
    return client.post('ventas/importar/', {
        archivo_base64,
        nombre_archivo: file.name,
        fecha: fecha || '',
    });
};
