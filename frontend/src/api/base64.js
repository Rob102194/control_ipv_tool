// Convierte un File/Blob a texto Base64 (sin el prefijo "data:...;base64,").
//
// Los archivos de importación (Excel) viajan en Base64 dentro del JSON, no
// como multipart/form-data: en la app de escritorio (Wails), el webview
// pierde el body de los POST con archivos binarios al pasar por el esquema
// wails:// (bug conocido de WKWebView, ver excel_handlers.go en el backend).
// Un JSON con el archivo en Base64 viaja igual que cualquier otro POST de la
// app, así que evitamos ese problema por completo.
export function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = reader.result;
      resolve(result.slice(result.indexOf(',') + 1));
    };
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}
