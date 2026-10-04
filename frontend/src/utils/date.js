// Formatea una fecha como YYYY-MM-DD usando el calendario local.
//
// No usar date.toISOString().split('T')[0]: toISOString() convierte a UTC,
// así que en husos horarios positivos (España) puede devolver el día
// anterior (o, al restar un día antes de convertir, dos días atrás).
export function formatDateLocal(date) {
    const pad = (n) => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

// Formatea una fecha "YYYY-MM-DD" en español ("12 de junio de 2026"),
// construyendo el Date con el calendario local (ver nota de formatDateLocal).
export function formatDateEs(fechaISO) {
    if (!fechaISO) return '';
    const [y, m, d] = fechaISO.split('-').map(Number);
    const date = new Date(y, m - 1, d);
    return date.toLocaleDateString('es-ES', { day: 'numeric', month: 'long', year: 'numeric' });
}

// Fecha de "ayer" (hoy - 1 día) en formato YYYY-MM-DD: el IPV se revisa a
// día vencido (un día después de que ocurrió la venta), así que es el valor
// por defecto natural al abrir Control IPV.
export function formatDateAyer() {
    const ayer = new Date();
    ayer.setDate(ayer.getDate() - 1);
    return formatDateLocal(ayer);
}
