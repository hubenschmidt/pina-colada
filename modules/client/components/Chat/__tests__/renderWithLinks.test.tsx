import { render } from '@testing-library/react';
import { renderWithLinks } from '../Chat';

// Mock the CSS module - must be defined inline since jest.mock() is hoisted
jest.mock('../Chat.module.css', () => ({
  errorText: 'errorText',
  link: 'link',
}));

const mockStyles = {
  errorText: 'errorText',
  link: 'link',
};

describe('renderWithLinks', () => {
  it('renders plain text correctly', () => {
    const result = renderWithLinks('Hello world');
    const { container } = render(<>{result}</>);
    
    expect(container.textContent).toBe('Hello world');
  });

  it('preserves newlines in text', () => {
    const text = 'Line 1\nLine 2\nLine 3';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const brElements = container.querySelectorAll('br');
    expect(brElements.length).toBe(2);
    expect(container.textContent).toContain('Line 1');
    expect(container.textContent).toContain('Line 2');
    expect(container.textContent).toContain('Line 3');
  });

  it('converts bare URLs to clickable links', () => {
    const text = 'Visit https://example.com for more info';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const link = container.querySelector('a');
    expect(link).toBeTruthy();
    expect(link?.href).toBe('https://example.com/');
    expect(link?.textContent).toBe('https://example.com');
  });

  it('converts markdown-style links to clickable links', () => {
    const text = 'Visit [Example](https://example.com) for more info';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const link = container.querySelector('a');
    expect(link).toBeTruthy();
    expect(link?.href).toBe('https://example.com/');
    expect(link?.textContent).toBe('Example');
  });

  it('styles error message in red', () => {
    const text = 'Some text\n\nSorry, there was an error generating the response.';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const errorSpan = container.querySelector(`.${mockStyles.errorText}`);
    expect(errorSpan).toBeTruthy();
    expect(errorSpan?.textContent).toBe('Sorry, there was an error generating the response.');
  });

  it('adds proper spacing before error message', () => {
    const text = 'Some response text\n\nSorry, there was an error generating the response.';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const brElements = container.querySelectorAll('br');
    // Should have at least 2 line breaks (one for the newline, one before error)
    expect(brElements.length).toBeGreaterThanOrEqual(2);
    
    const errorSpan = container.querySelector(`.${mockStyles.errorText}`);
    expect(errorSpan).toBeTruthy();
  });

  it('handles error message with links in preceding text', () => {
    const text = 'Visit https://example.com\n\nSorry, there was an error generating the response.';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const link = container.querySelector('a');
    expect(link).toBeTruthy();
    
    const errorSpan = container.querySelector(`.${mockStyles.errorText}`);
    expect(errorSpan).toBeTruthy();
    expect(errorSpan?.textContent).toBe('Sorry, there was an error generating the response.');
  });

  it('handles multiple links and error message', () => {
    const text = 'Check https://site1.com and https://site2.com\n\nSorry, there was an error generating the response.';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const links = container.querySelectorAll('a');
    expect(links.length).toBe(2);
    
    const errorSpan = container.querySelector(`.${mockStyles.errorText}`);
    expect(errorSpan).toBeTruthy();
  });

  it('handles empty text', () => {
    const result = renderWithLinks('');
    expect(result).toEqual([]);
  });

  it('handles text with only error message', () => {
    const text = 'Sorry, there was an error generating the response.';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const errorSpan = container.querySelector(`.${mockStyles.errorText}`);
    expect(errorSpan).toBeTruthy();
    expect(errorSpan?.textContent).toBe('Sorry, there was an error generating the response.');
  });

  it('handles error message in middle of text with links', () => {
    const text = 'Before https://example.com\n\nSorry, there was an error generating the response.\nAfter';
    const result = renderWithLinks(text);
    const { container } = render(<>{result}</>);
    
    const link = container.querySelector('a');
    expect(link).toBeTruthy();
    
    const errorSpan = container.querySelector(`.${mockStyles.errorText}`);
    expect(errorSpan).toBeTruthy();
    
    expect(container.textContent).toContain('Before');
    expect(container.textContent).toContain('After');
  });
});

