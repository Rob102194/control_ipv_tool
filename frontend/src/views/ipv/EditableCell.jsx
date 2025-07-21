import React, { useState, useEffect } from 'react';
import { Form, Button, OverlayTrigger, Tooltip } from 'react-bootstrap';
import CommentModal from './CommentModal';

function EditableCell({ value, onChange, onCommentChange, comment }) {
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
                // eslint-disable-next-line no-eval
                const result = eval(inputValue);
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
                />
            )}
            <Button variant={comment ? "info" : "link"} size="sm" onClick={() => setShowCommentModal(true)}>
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
