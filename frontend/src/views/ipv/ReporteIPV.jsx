import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { formatDateEs } from '../../utils/date';

// RGB de la paleta "Aire Alpino" (ver :root en frontend/src/index.css).
// jsPDF no entiende variables CSS, así que se repiten aquí como constantes
// para que el PDF se sienta parte de la misma app en vez de un documento
// genérico aparte.
const PRIMARY = [26, 50, 156];      // --color-primary
const POSITIVO = [6, 173, 129];     // --color-tertiary (sobrante)
const NEGATIVO = [199, 19, 61];     // --error (faltante)
const NEUTRO = [100, 116, 139];     // --text-secondary (diferencia en cero)
const MERMA_INK = [122, 93, 21];    // --warning-ink
const TEXTO = [30, 41, 59];         // --text-main
const BORDE = [226, 232, 240];      // --border-color
const ZEBRA = [248, 250, 252];      // --color-base

const MARGIN_X = 14;

// Colorea el texto de una diferencia igual que el ipv-badge en pantalla:
// verde si sobra, rojo si falta, gris si cierra en cero.
function colorDiferencia(valorCelda) {
    const n = parseFloat(valorCelda);
    if (n < 0) return NEGATIVO;
    if (n > 0) return POSITIVO;
    return NEUTRO;
}

const tableTheme = {
    styles: { fontSize: 8.5, textColor: TEXTO, lineColor: BORDE },
    headStyles: { fillColor: PRIMARY, textColor: 255, fontStyle: 'bold', fontSize: 8.5 },
    alternateRowStyles: { fillColor: ZEBRA },
    margin: { left: MARGIN_X, right: MARGIN_X },
};

const ReporteIPV = (reporteData) => {
    const doc = new jsPDF();
    const pageWidth = doc.internal.pageSize.width;
    const pageHeight = doc.internal.pageSize.height;
    let y = 20;

    const checkPageBreak = (currentY, minSpace = 24) => {
        if (currentY >= pageHeight - minSpace) {
            doc.addPage();
            return 20;
        }
        return currentY;
    };

    // --- Encabezado ---
    doc.setFont(undefined, 'bold');
    doc.setFontSize(9);
    doc.setTextColor(...PRIMARY);
    doc.text('CONTROL IPV', MARGIN_X, y);
    y += 8;

    doc.setFontSize(18);
    doc.setTextColor(...TEXTO);
    doc.text('Reporte de Inventario Diario', MARGIN_X, y);
    y += 7;

    doc.setFont(undefined, 'normal');
    doc.setFontSize(11);
    doc.setTextColor(...NEUTRO);
    doc.text(formatDateEs(reporteData.fecha), MARGIN_X, y);
    y += 5;

    doc.setDrawColor(...BORDE);
    doc.line(MARGIN_X, y, pageWidth - MARGIN_X, y);
    y += 9;

    // --- Tablas por área ---
    const columnasArea = ["Producto", "UM", "Inicio", "Entradas", "Consumo", "Merma", "O.S.", "F. Teórico", "F. Físico", "Diferencia"];
    const alinearDerecha = {};
    for (let i = 2; i <= 9; i++) alinearDerecha[i] = { halign: 'right' };
    alinearDerecha[9] = { halign: 'right', fontStyle: 'bold' };

    for (const areaNombre in reporteData.areas) {
        y = checkPageBreak(y, 34);
        doc.setFont(undefined, 'bold');
        doc.setFontSize(13);
        doc.setTextColor(...PRIMARY);
        doc.text(areaNombre, MARGIN_X, y);
        y += 6;

        const filas = reporteData.areas[areaNombre].map(item => ([
            item.producto, item.um, item.inicio, item.entradas, item.consumo,
            item.merma, item.otras_salidas, item.final_teorico, item.final_fisico, item.diferencia,
        ]));

        autoTable(doc, {
            ...tableTheme,
            head: [columnasArea],
            body: filas,
            startY: y,
            columnStyles: alinearDerecha,
            didParseCell: (data) => {
                if (data.section === 'body' && data.column.index === 9) {
                    data.cell.styles.textColor = colorDiferencia(data.cell.raw);
                }
            },
        });
        y = doc.lastAutoTable.finalY + 10;
    }

    // --- Resumen por área (tabla, no texto suelto) ---
    const areasConResumen = Object.entries(reporteData.resumen).filter(
        ([, r]) => r.faltantes.length > 0 || r.sobrantes.length > 0 || r.mermas.length > 0
    );
    if (areasConResumen.length > 0) {
        y = checkPageBreak(y, 34);
        doc.setFont(undefined, 'bold');
        doc.setFontSize(14);
        doc.setTextColor(...TEXTO);
        doc.text('Resumen por área', MARGIN_X, y);
        y += 7;

        const filasResumen = [];
        areasConResumen.forEach(([areaNombre, r]) => {
            r.faltantes.forEach(f => filasResumen.push([areaNombre, 'Faltante', f.producto, `${f.cantidad} ${f.um}`]));
            r.sobrantes.forEach(s => filasResumen.push([areaNombre, 'Sobrante', s.producto, `${s.cantidad} ${s.um}`]));
            r.mermas.forEach(m => filasResumen.push([areaNombre, 'Merma', m.producto, `${m.cantidad} ${m.um}`]));
        });

        autoTable(doc, {
            ...tableTheme,
            head: [['Área', 'Tipo', 'Producto', 'Cantidad']],
            body: filasResumen,
            startY: y,
            styles: { ...tableTheme.styles, fontSize: 9 },
            columnStyles: { 1: { cellWidth: 26, fontStyle: 'bold' }, 3: { halign: 'right' } },
            didParseCell: (data) => {
                if (data.section === 'body' && data.column.index === 1) {
                    const tipo = data.cell.raw;
                    if (tipo === 'Faltante') data.cell.styles.textColor = NEGATIVO;
                    else if (tipo === 'Sobrante') data.cell.styles.textColor = POSITIVO;
                    else data.cell.styles.textColor = MERMA_INK;
                }
            },
        });
        y = doc.lastAutoTable.finalY + 10;
    }

    // --- Notas ---
    const hayNotas = Object.values(reporteData.notas).some(arr => arr.length > 0);
    if (hayNotas) {
        y = checkPageBreak(y, 30);
        doc.setFont(undefined, 'bold');
        doc.setFontSize(14);
        doc.setTextColor(...TEXTO);
        doc.text('Notas y comentarios', MARGIN_X, y);
        y += 8;

        for (const areaNombre in reporteData.notas) {
            if (reporteData.notas[areaNombre].length === 0) continue;
            y = checkPageBreak(y, 20);
            doc.setFont(undefined, 'bold');
            doc.setFontSize(11);
            doc.setTextColor(...PRIMARY);
            doc.text(areaNombre, MARGIN_X, y);
            y += 6;

            doc.setFont(undefined, 'normal');
            doc.setFontSize(9.5);
            doc.setTextColor(...TEXTO);
            reporteData.notas[areaNombre].forEach(nota => {
                y = checkPageBreak(y, 16);
                const lineas = doc.splitTextToSize(`•  ${nota}`, pageWidth - MARGIN_X * 2 - 4);
                doc.text(lineas, MARGIN_X + 2, y);
                y += lineas.length * 5;
            });
            y += 4;
        }
    }

    // --- Pie de página: marca + numeración, en todas las páginas ---
    const totalPaginas = doc.internal.getNumberOfPages();
    const generadoEl = new Date().toLocaleString('es-ES', { dateStyle: 'short', timeStyle: 'short' });
    for (let i = 1; i <= totalPaginas; i++) {
        doc.setPage(i);
        doc.setDrawColor(...BORDE);
        doc.line(MARGIN_X, pageHeight - 14, pageWidth - MARGIN_X, pageHeight - 14);
        doc.setFont(undefined, 'normal');
        doc.setFontSize(8);
        doc.setTextColor(...NEUTRO);
        doc.text(`Control IPV · generado el ${generadoEl}`, MARGIN_X, pageHeight - 9);
        doc.text(`Página ${i} de ${totalPaginas}`, pageWidth - MARGIN_X, pageHeight - 9, { align: 'right' });
    }

    doc.save(`reporte_ipv_${reporteData.fecha}.pdf`);
};

export default ReporteIPV;
