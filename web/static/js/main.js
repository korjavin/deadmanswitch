/**
 * Dead Man's Switch - Main JavaScript
 */

document.addEventListener('DOMContentLoaded', () => {
  // Initialize components
  initAlerts();
  initDropdowns();
  initCopyButtons();
  initPasswordToggles();
  initTabSwitching();
  initCountdowns();
});

/**
 * Alert message handling
 */
function initAlerts() {
  // Close alert when the close button is clicked
  document.querySelectorAll('.alert-close').forEach(button => {
    button.addEventListener('click', () => {
      const alert = button.closest('.alert');
      if (alert) {
        alert.style.opacity = '0';
        setTimeout(() => {
          alert.style.display = 'none';
        }, 300);
      }
    });
  });

  // Auto-hide success alerts after 5 seconds
  document.querySelectorAll('.alert-success').forEach(alert => {
    setTimeout(() => {
      alert.style.opacity = '0';
      setTimeout(() => {
        alert.style.display = 'none';
      }, 300);
    }, 5000);
  });
}

/**
 * Dropdown menu handling
 */
function initDropdowns() {
  // For mobile: Show/hide dropdowns on click instead of hover
  if (window.innerWidth < 768) {
    document.querySelectorAll('.dropdown-toggle').forEach(toggle => {
      toggle.addEventListener('click', (e) => {
        e.preventDefault();
        const dropdown = toggle.nextElementSibling;
        if (dropdown.style.display === 'block') {
          dropdown.style.display = 'none';
        } else {
          // Hide all other dropdowns
          document.querySelectorAll('.dropdown-menu').forEach(menu => {
            menu.style.display = 'none';
          });
          dropdown.style.display = 'block';
        }
      });
    });

    // Close dropdowns when clicking outside
    document.addEventListener('click', (e) => {
      if (!e.target.closest('.dropdown')) {
        document.querySelectorAll('.dropdown-menu').forEach(menu => {
          menu.style.display = 'none';
        });
      }
    });
  }
}

/**
 * Copy to clipboard functionality
 */
function initCopyButtons() {
  document.querySelectorAll('.copy-btn').forEach(button => {
    button.addEventListener('click', () => {
      const target = document.getElementById(button.dataset.target);
      if (!target) return;

      // Select the text
      if (target.tagName.toLowerCase() === 'input' || target.tagName.toLowerCase() === 'textarea') {
        target.select();
        target.setSelectionRange(0, 99999); // For mobile devices
      } else {
        const range = document.createRange();
        range.selectNode(target);
        window.getSelection().removeAllRanges();
        window.getSelection().addRange(range);
      }

      try {
        // Copy the text
        document.execCommand('copy');
        
        // Show feedback
        const originalText = button.textContent;
        button.textContent = 'Copied!';
        button.classList.add('btn-success');
        button.classList.remove('btn-secondary');
        
        // Reset button after 2 seconds
        setTimeout(() => {
          button.textContent = originalText;
          button.classList.remove('btn-success');
          button.classList.add('btn-secondary');
        }, 2000);
      } catch (err) {
        console.error('Failed to copy text: ', err);
      }
      
      // Clear selection
      window.getSelection().removeAllRanges();
    });
  });
}

/**
 * Password visibility toggle
 */
function initPasswordToggles() {
  document.querySelectorAll('.password-toggle').forEach(toggle => {
    toggle.addEventListener('click', () => {
      const passwordInput = document.getElementById(toggle.dataset.target);
      if (!passwordInput) return;
      
      if (passwordInput.type === 'password') {
        passwordInput.type = 'text';
        toggle.textContent = 'Hide';
      } else {
        passwordInput.type = 'password';
        toggle.textContent = 'Show';
      }
    });
  });
}

/**
 * Tab switching for multi-tab content
 */
function initTabSwitching() {
  document.querySelectorAll('.tab-link').forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      
      const targetId = link.getAttribute('href').substring(1);
      const tabContent = document.getElementById(targetId);
      if (!tabContent) return;
      
      // Hide all tab content
      document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
      });
      
      // Show target tab content
      tabContent.classList.add('active');
      
      // Update active tab link
      document.querySelectorAll('.tab-link').forEach(tabLink => {
        tabLink.classList.remove('active');
      });
      link.classList.add('active');
    });
  });
}

/**
 * Countdown timers for deadlines
 */
function initCountdowns() {
  document.querySelectorAll('.countdown').forEach(element => {
    const deadline = new Date(element.dataset.deadline).getTime();
    
    // Update every second
    const interval = setInterval(() => {
      const now = new Date().getTime();
      const distance = deadline - now;
      
      // If countdown is finished
      if (distance < 0) {
        clearInterval(interval);
        element.innerHTML = "EXPIRED";
        element.classList.add('countdown-expired');
        return;
      }
      
      // Calculate time units
      const days = Math.floor(distance / (1000 * 60 * 60 * 24));
      const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
      const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((distance % (1000 * 60)) / 1000);
      
      // Display the countdown
      element.innerHTML = `${days}d ${hours}h ${minutes}m ${seconds}s`;
      
      // Add warning class when there's less than a day left
      if (days === 0) {
        element.classList.add('countdown-warning');
      }
    }, 1000);
  });
}

/**
 * Handle form submissions with AJAX
 */
function ajaxSubmit(formElement, successCallback, errorCallback) {
  formElement.addEventListener('submit', (e) => {
    e.preventDefault();
    
    const formData = new FormData(formElement);
    const url = formElement.getAttribute('action');
    const method = formElement.getAttribute('method') || 'POST';
    
    // Show loading state
    const submitButton = formElement.querySelector('button[type="submit"]');
    const originalText = submitButton ? submitButton.textContent : '';
    if (submitButton) {
      submitButton.disabled = true;
      submitButton.textContent = 'Processing...';
    }
    
    // Send the request
    fetch(url, {
      method: method,
      body: formData,
      headers: {
        'X-Requested-With': 'XMLHttpRequest'
      }
    })
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP error ${response.status}`);
      }
      return response.json();
    })
    .then(data => {
      if (typeof successCallback === 'function') {
        successCallback(data);
      }
    })
    .catch(error => {
      console.error('Error submitting form:', error);
      if (typeof errorCallback === 'function') {
        errorCallback(error);
      }
    })
    .finally(() => {
      // Reset button state
      if (submitButton) {
        submitButton.disabled = false;
        submitButton.textContent = originalText;
      }
    });
  });
}

/**
 * Format a date for display
 */
function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleString();
}

/**
 * Show a modal dialog
 */
function showModal(id) {
  const modal = document.getElementById(id);
  if (modal) {
    modal.style.display = 'block';
    setTimeout(() => {
      modal.classList.add('show');
    }, 10);
  }
}

/**
 * Hide a modal dialog
 */
function hideModal(id) {
  const modal = document.getElementById(id);
  if (modal) {
    modal.classList.remove('show');
    setTimeout(() => {
      modal.style.display = 'none';
    }, 300);
  }
}

/**
 * Initialize modal handlers
 */
function initModals() {
  // Open modals
  document.querySelectorAll('[data-modal]').forEach(trigger => {
    trigger.addEventListener('click', () => {
      const modalId = trigger.dataset.modal;
      showModal(modalId);
    });
  });
  
  // Close modals with close buttons
  document.querySelectorAll('.modal-close').forEach(closeButton => {
    closeButton.addEventListener('click', () => {
      const modal = closeButton.closest('.modal');
      if (modal) {
        hideModal(modal.id);
      }
    });
  });
  
  // Close modals when clicking on backdrop
  document.querySelectorAll('.modal').forEach(modal => {
    modal.addEventListener('click', (e) => {
      if (e.target === modal) {
        hideModal(modal.id);
      }
    });
  });
  
  // Close modals with Escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      document.querySelectorAll('.modal.show').forEach(modal => {
        hideModal(modal.id);
      });
    }
  });
}