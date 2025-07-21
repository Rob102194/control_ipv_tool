import client from './client';

export const getVentas = () => client.get('ventas/');
export const updateVenta = (id, data) => client.put(`ventas/${id}/`, data);
export const deleteVenta = (id) => client.delete(`ventas/${id}/`);
export const deleteVentas = (ids) => client.post('ventas/delete-multiple/', { ids });
export const importVentas = (file, fecha) => {
    const formData = new FormData();
    formData.append('file', file);
    if (fecha) {
        formData.append('fecha', fecha);
    }
    return client.post('ventas/importar/', formData, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    });
};
