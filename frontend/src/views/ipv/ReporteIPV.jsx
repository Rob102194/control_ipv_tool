import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';

const ReporteIPV = (reporteData) => {
    const doc = new jsPDF();
    const pageHeight = doc.internal.pageSize.height;
    let y = 30;

    const checkPageBreak = (currentY) => {
        if (currentY >= pageHeight - 20) {
            doc.addPage();
            return 22; // Reiniciar y en la nueva página
        }
        return currentY;
    };

    // Título
    doc.setFontSize(18);
    doc.text(`Reporte de IPV - ${reporteData.fecha}`, 14, 22);

    // Tablas por área
    for (const areaNombre in reporteData.areas) {
        y = checkPageBreak(y);
        doc.setFontSize(14);
        doc.text(areaNombre, 14, y);
        y += 7;

        const tableColumn = ["Producto", "UM", "Inicio", "Entradas", "Consumo", "Merma", "O.S.", "F. Teórico", "F. Físico", "Diferencia"];
        const tableRows = [];

        reporteData.areas[areaNombre].forEach(item => {
            const rowData = [
                item.producto,
                item.um,
                item.inicio,
                item.entradas,
                item.consumo,
                item.merma,
                item.otras_salidas,
                item.final_teorico,
                item.final_fisico,
                item.diferencia
            ];
            tableRows.push(rowData);
        });

        autoTable(doc, {
            head: [tableColumn],
            body: tableRows,
            startY: y,
            didDrawPage: (data) => {
                y = data.cursor.y + 10;
            }
        });
        y = doc.lastAutoTable.finalY + 10;
    }

    // Resumen por Área
    y = checkPageBreak(y);
    doc.setFontSize(14);
    doc.text("Resumen por Área", 14, y);
    y += 7;

    for (const areaNombre in reporteData.resumen) {
        y = checkPageBreak(y);
        const resumenArea = reporteData.resumen[areaNombre];
        if (resumenArea.faltantes.length === 0 && resumenArea.sobrantes.length === 0 && resumenArea.mermas.length === 0) {
            continue;
        }

        doc.setFontSize(12);
        doc.text(areaNombre, 14, y);
        y += 6;
        doc.setFontSize(10);

        if (resumenArea.faltantes.length > 0) {
            y = checkPageBreak(y);
            doc.text("Faltantes:", 16, y);
            y += 5;
            resumenArea.faltantes.forEach(item => {
                y = checkPageBreak(y);
                doc.text(`- ${item}`, 18, y);
                y += 5;
            });
        }

        if (resumenArea.sobrantes.length > 0) {
            y = checkPageBreak(y);
            doc.text("Sobrantes:", 16, y);
            y += 5;
            resumenArea.sobrantes.forEach(item => {
                y = checkPageBreak(y);
                doc.text(`- ${item}`, 18, y);
                y += 5;
            });
        }

        if (resumenArea.mermas.length > 0) {
            y = checkPageBreak(y);
            doc.text("Mermas:", 16, y);
            y += 5;
            resumenArea.mermas.forEach(item => {
                y = checkPageBreak(y);
                doc.text(`- ${item}`, 18, y);
                y += 5;
            });
        }
        y += 5; // Espacio entre áreas
    }
    
    // Notas
    if (Object.keys(reporteData.notas).length > 0) {
        y = checkPageBreak(y + 10);
        doc.setFontSize(14);
        doc.text("Notas y Comentarios", 14, y);
        y += 7;

        for (const areaNombre in reporteData.notas) {
            if (reporteData.notas[areaNombre].length > 0) {
                y = checkPageBreak(y);
                doc.setFontSize(12);
                doc.text(areaNombre, 14, y);
                y += 6;
                doc.setFontSize(10);
                reporteData.notas[areaNombre].forEach(nota => {
                    y = checkPageBreak(y);
                    const splitNota = doc.splitTextToSize(nota, 180);
                    doc.text(splitNota, 16, y);
                    y += (splitNota.length * 5);
                });
                y += 5; // Espacio entre áreas
            }
        }
    }


    doc.save(`reporte_ipv_${reporteData.fecha}.pdf`);
};

export default ReporteIPV;
