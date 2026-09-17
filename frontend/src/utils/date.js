// Formatea una fecha como YYYY-MM-DD usando el calendario local.
//
// No usar date.toISOString().split('T')[0]: toISOString() convierte a UTC,
// así que en husos horarios positivos (España) puede devolver el día
// anterior (o, al restar un día antes de convertir, dos días atrás).
export function formatDateLocal(date) {
    const pad = (n) => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}
