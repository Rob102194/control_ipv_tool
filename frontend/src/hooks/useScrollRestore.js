import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

// Restaura la posición de scroll de una lista al volver de editar un
// elemento (p. ej. entrar a "Editar" y volver con Cancelar/Guardar/"atrás").
// `prefix` distingue la lista (p. ej. "recetas", "productos"); `listo` debe
// ser true solo cuando los datos ya terminaron de cargar y la lista tiene su
// altura real.
//
// `location.key` identifica esta entrada del historial (la misma al volver
// con "atrás"/navigate(-1), nueva en cada navegación fresca), así que sirve
// de clave para guardar/recuperar el scroll.
//
// No basta con leer window.scrollY al desmontar: para cuando React desmonta
// la lista y monta el formulario, el documento ya se achicó y el propio
// navegador ya recortó el scroll a 0 — se captura DEMASIADO TARDE. Por eso
// se guarda en cada scroll, mientras la lista sigue siendo la página
// completa, y se restaura con behavior:'instant' (el proyecto fija
// scroll-behavior:smooth a nivel global, y con listas largas una animación
// "suave" tardaría varios segundos en lugar de ser instantánea).
export function useScrollRestore(prefix, listo) {
    const location = useLocation();
    const scrollKey = `${prefix}-scroll:${location.key}`;

    useEffect(() => {
        let frame = null;
        const onScroll = () => {
            if (frame) return;
            frame = requestAnimationFrame(() => {
                sessionStorage.setItem(scrollKey, String(window.scrollY));
                frame = null;
            });
        };
        window.addEventListener('scroll', onScroll);
        return () => {
            window.removeEventListener('scroll', onScroll);
            if (frame) cancelAnimationFrame(frame);
        };
    }, [scrollKey]);

    useEffect(() => {
        if (!listo) return;
        const guardado = sessionStorage.getItem(scrollKey);
        if (guardado) {
            requestAnimationFrame(() => window.scrollTo({ top: parseInt(guardado, 10), behavior: 'instant' }));
        }
    }, [listo, scrollKey]);
}
