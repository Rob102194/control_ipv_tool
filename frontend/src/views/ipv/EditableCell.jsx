import React, { useState, useEffect } from 'react';
import { Form, Button, OverlayTrigger, Tooltip } from 'react-bootstrap';
import CommentModal from './CommentModal';

// Evalúa expresiones aritméticas simples (+ - * / y paréntesis) sin usar
// eval(): evita ejecutar JS arbitrario y el bug de los literales octales
// legacy de JS (eval("013") da 11, no 13).
function evaluarExpresion(expr) {
    let i = 0;
    const skipSpaces = () => { while (expr[i] === ' ') i++; };
    const parseNumber = () => {
        skipSpaces();
        const start = i;
        if (expr[i] === '+' || expr[i] === '-') i++;
        const digitsStart = i;
        while (i < expr.length && /[0-9.]/.test(expr[i])) i++;
        if (i === digitsStart) throw new Error('número esperado');
        return parseFloat(expr.slice(start, i));
    };
    const parseFactor = () => {
        skipSpaces();
        if (expr[i] === '(') {
            i++;
            const value = parseExpr();
            skipSpaces();
            if (expr[i] !== ')') throw new Error('paréntesis sin cerrar');
            i++;
            return value;
        }
        return parseNumber();
    };
    const parseTerm = () => {
        let value = parseFactor();
        skipSpaces();
        while (expr[i] === '*' || expr[i] === '/') {
            const op = expr[i]; i++;
            const rhs = parseFactor();
            value = op === '*' ? value * rhs : value / rhs;
            skipSpaces();
        }
        return value;
    };
    const parseExpr = () => {
        let value = parseTerm();
        skipSpaces();
        while (expr[i] === '+' || expr[i] === '-') {
            const op = expr[i]; i++;
            const rhs = parseTerm();
            value = op === '+' ? value + rhs : value - rhs;
            skipSpaces();
        }
        return value;
    };
    const result = parseExpr();
    skipSpaces();
    if (i !== expr.length) throw new Error('expresión inválida');
    return result;
}

function EditableCell({ value, onChange, onCommentChange, comment, label }) {
    const [inputValue, setInputValue] = useState(value);
    const [showCommentModal, setShowCommentModal] = useState(false);

    useEffect(() => {
        setInputValue(value);
    }, [value]);

    const handleFocus = (e) => {
        if (parseFloat(e.target.value) === 0) {
            setInputValue('');
        }
    };

    const handleBlur = (e) => {
        if (e.target.value === '') {
            setInputValue(0);
            onChange(0);
        } else {
            onChange(e.target.value);
        }
    };

    const handleKeyDown = (e) => {
        if (e.key === 'Enter') {
            try {
                const result = evaluarExpresion(inputValue);
                if (!isNaN(result)) {
                    const fixedResult = parseFloat(result).toFixed(3);
                    setInputValue(fixedResult);
                    onChange(fixedResult);
                }
            } catch (error) {
                // Ignore evaluation errors
            }
        }
    };

    const handleChange = (e) => {
        setInputValue(e.target.value);
    };

    const handleSaveComment = (text) => {
        onCommentChange(text);
    };

    const handleDeleteComment = () => {
        onCommentChange('');
    };

    const renderTooltip = (props) => (
        <Tooltip id="comment-tooltip" {...props}>
            {comment}
        </Tooltip>
    );

    return (
        <div style={{ display: 'flex', alignItems: 'center' }}>
            {comment ? (
                <OverlayTrigger
                    placement="top"
                    overlay={renderTooltip}
                >
                    <Form.Control
                        type="text"
                        value={inputValue}
                        onFocus={handleFocus}
                        onBlur={handleBlur}
                        onChange={handleChange}
                        onKeyDown={handleKeyDown}
                        min="0"
                        aria-label={label}
                    />
                </OverlayTrigger>
            ) : (
                <Form.Control
                    type="text"
                    value={inputValue}
                    onFocus={handleFocus}
                    onBlur={handleBlur}
                    onChange={handleChange}
                    onKeyDown={handleKeyDown}
                    min="0"
                    aria-label={label}
                />
            )}
            <Button
                variant={comment ? "info" : "link"}
                size="sm"
                onClick={() => setShowCommentModal(true)}
                aria-label={comment ? `Editar comentario de ${label}` : `Agregar comentario a ${label}`}
            >
                ...
            </Button>
            <CommentModal
                show={showCommentModal}
                onHide={() => setShowCommentModal(false)}
                onSave={handleSaveComment}
                onDelete={handleDeleteComment}
                comment={comment}
            />
        </div>
    );
}

export default EditableCell;
