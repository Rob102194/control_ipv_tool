# Control de Inventario para Restaurantes (Control IPV)

## Introducción

**Control IPV** es una aplicación de escritorio diseñada para simplificar la gestión del inventario diario en restaurantes y negocios de hostelería. Permite llevar un control preciso de productos, recetas, ventas y áreas de trabajo, ayudando a identificar sobrantes y faltantes para optimizar la gestión de costos.

La aplicación cuenta con una interfaz de usuario moderna y amigable, con funcionalidades como modo claro/oscuro, y está empaquetada como un ejecutable único para facilitar su distribución y uso en sistemas Windows.

---

## Características Principales

- **Gestión de Productos**: Cree y administre una lista de todos los productos e ingredientes utilizados en su negocio.
- **Gestión de Áreas**: Defina diferentes áreas de trabajo (ej. Barra, Cocina) para organizar su inventario.
- **Creación de Recetas**: Especifique los ingredientes y las cantidades para cada plato, permitiendo un cálculo de consumo preciso.
- **Control de Inventario Diario (IPV)**:
    - Registre el inventario inicial, entradas, mermas y otras salidas.
    - Calcule automáticamente el consumo basado en las ventas del día.
    - Determine el inventario final teórico y compárelo con el conteo físico para identificar diferencias.
- **Reportes Detallados**: Genere reportes en PDF del estado del inventario, incluyendo resúmenes de sobrantes, faltantes y notas importantes.
- **Importación y Exportación de Datos**: Agilice la configuración inicial y la gestión diaria importando y exportando productos y recetas desde y hacia archivos Excel.
- **Modo Claro y Oscuro**: Cambie entre temas para una mejor visualización según sus preferencias.
- **Aplicación de Escritorio**: Funciona como una aplicación nativa en Windows, con un ícono en la barra de tareas para un acceso rápido.

---

## ¿Cómo Usar la Aplicación? (Para Usuarios Finales)

Para utilizar la aplicación, no necesita instalar nada. Simplemente siga estos pasos:

1.  **Descargue la Carpeta**: Obtenga la carpeta que contiene la aplicación (por defecto llamada `dist`, pero puede tener otro nombre).
2.  **Ejecute el Programa**: Dentro de la carpeta, haga doble clic en el archivo `Control IPV.exe`.
3.  **Base de Datos**: La primera vez que ejecute el programa, se creará automáticamente un archivo llamado `inventario.db` en la misma carpeta. **¡No borre este archivo!** Contiene toda la información de sus productos, recetas e inventarios.
4.  **Acceso Rápido**: La aplicación se mantendrá corriendo en segundo plano. Puede acceder a ella a través del ícono que aparecerá en la barra de tareas (cerca del reloj).

---

## Importación de Datos

La aplicación permite importar registros de **Productos**, **Recetas** y **Ventas** directamente desde archivos Excel, lo que agiliza enormemente el proceso de carga de datos.

### Formato para Importar Productos

Cree un archivo Excel (`.xlsx`) con las siguientes columnas:

-   `nombre`: El nombre único del producto.
-   `unidad_medida`: La unidad en la que se mide el producto (ej. `kg`, `l`, `unidades`).

**Ejemplo:**

| nombre      | unidad_medida |
| ----------- | ------------- |
| Queso Gouda | kg            |
| Tomate      | kg            |
| Pan de Papa | unidades      |

### Formato para Importar Recetas

Cree un archivo Excel (`.xlsx`) con las siguientes columnas para definir las recetas y sus ingredientes:

-   `receta_nombre`: El nombre de la receta. Repita el nombre para cada ingrediente que le pertenece.
-   `producto_nombre`: El nombre del ingrediente.
-   `unidad_medida`: La unidad de medida del ingrediente.
-   `cantidad`: La cantidad del ingrediente utilizada en la receta.
-   `area_nombre`: El nombre del área de donde se descuenta el ingrediente.

**Importante**: Si un producto o área no existe, la aplicación lo creará automáticamente.

**Ejemplo:**

| receta_nombre     | producto_nombre | unidad_medida | cantidad | area_nombre |
| ----------------- | --------------- | ------------- | -------- | ----------- |
| Hamburguesa Clásica | Pan de Papa     | unidades      | 1        | Cocina      |
| Hamburguesa Clásica | Carne de Res    | kg            | 0.15     | Cocina      |
| Hamburguesa Clásica | Queso Gouda     | kg            | 0.02     | Cocina      |

### Formato para Importar Ventas

Cree un archivo Excel (`.xlsx`) con las siguientes columnas:

-   `Nombre`: El nombre exacto de la receta vendida.
-   `Cantidad`: El número de unidades vendidas.

**Importante**: Si una receta no existe, se creará automáticamente (sin ingredientes) para que pueda editarla más tarde.

**Ejemplo:**

| Nombre            | Cantidad |
| ----------------- | -------- |
| Hamburguesa Clásica | 15       |
| Refresco de Cola  | 25       |

---

## Para Desarrolladores

Esta sección contiene instrucciones para aquellos que deseen modificar o contribuir al código fuente.

### Estructura del Proyecto

El proyecto está dividido en dos partes principales:

-   `backend/`: Una aplicación Flask (Python) que gestiona la lógica de negocio y la base de datos.
-   `frontend/`: Una aplicación de React (JavaScript) que construye la interfaz de usuario.

### Prerrequisitos

-   Node.js y npm (para el frontend)
-   Python y pip (para el backend)

### Configuración del Entorno

1.  **Clonar el Repositorio**:
    ```bash
    git clone <URL_DEL_REPOSITORIO>
    cd inventario-restaurante
    ```

2.  **Configurar el Backend**:
    ```bash
    # Navegar a la carpeta del backend
    cd backend

    # Crear y activar un entorno virtual
    python -m venv venv
    venv\Scripts\activate

    # Instalar las dependencias
    pip install -r requirements.txt
    ```

3.  **Configurar el Frontend**:
    ```bash
    # Navegar a la carpeta del frontend
    cd ../frontend

    # Instalar las dependencias
    npm install
    ```

### Ejecutar en Modo de Desarrollo

Para ver los cambios en tiempo real, debe ejecutar ambos servidores simultáneamente en terminales separadas:

-   **Iniciar Backend**:
    ```bash
    # Desde la carpeta raíz del proyecto
    cd backend
    venv\Scripts\activate
    python app.py
    ```

-   **Iniciar Frontend**:
    ```bash
    # Desde la carpeta raíz del proyecto
    cd frontend
    npm run dev
    ```

### Construir el Ejecutable para Producción

Para generar el archivo `.exe` final, siga estos pasos:

1.  **Construir el Frontend**:
    ```bash
    cd frontend
    npm run build
    ```

2.  **Construir el Ejecutable con PyInstaller**:
    ```bash
    # Desde la carpeta raíz del proyecto
    cd backend
    venv\Scripts\activate
    cd ..
    pyinstaller "Control IPV.spec"
    ```
    O usando el comando directo:
    ```bash
    backend\venv\Scripts\activate && python -m PyInstaller --noconfirm --onefile --windowed --icon="frontend/public/icon.png" --name "Control IPV" --add-data "frontend/dist;frontend/dist" --distpath "./dist" --hidden-import=dependency_injector.errors backend/app.py
    ```

El ejecutable final y todos los archivos necesarios se encontrarán en la carpeta `dist`.

---

