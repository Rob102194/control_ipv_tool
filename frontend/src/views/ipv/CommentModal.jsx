import React, { useState, useEffect } from 'react';
import { Modal, Button, Form } from 'react-bootstrap';

function CommentModal({ show, onHide, onSave, onDelete, comment }) {
    const [text, setText] = useState(comment || '');

    useEffect(() => {
        setText(comment || '');
    }, [comment, show]);

    const handleSave = () => {
        onSave(text);
        onHide();
    };

    const handleDelete = () => {
        onDelete();
        onHide();
    };

    return (
        <Modal show={show} onHide={onHide}>
            <Modal.Header closeButton>
                <Modal.Title>Comentario</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form.Control
                    as="textarea"
                    rows={3}
                    value={text}
                    onChange={(e) => setText(e.target.value)}
                />
            </Modal.Body>
            <Modal.Footer>
                <Button variant="danger" onClick={handleDelete}>Eliminar</Button>
                <Button variant="secondary" onClick={onHide}>
                    Cerrar
                </Button>
                <Button variant="primary" onClick={handleSave}>
                    Guardar
                </Button>
            </Modal.Footer>
        </Modal>
    );
}

export default CommentModal;
