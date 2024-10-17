import React from 'react';

interface InputProps {
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  className?: string;
}

const Input: React.FC<InputProps> = ({ value, onChange, placeholder = '', className = '' }) => (
  <input
    type='text'
    value={value}
    onChange={onChange}
    placeholder={placeholder}
    className={`px-3 py-2 border rounded ${className}`}
  />
);

export default Input;
