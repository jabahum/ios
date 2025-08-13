// Lab Sample Types Dynamic Dropdowns
document.addEventListener('DOMContentLoaded', function() {
    // Initialize the lab sample type functionality
    initializeLabSampleTypes();
});

function initializeLabSampleTypes() {
    // Add event listeners to sample type checkboxes
    const sampleBloodCheckbox = document.getElementById('sample_blood');
    const sampleUrineCheckbox = document.getElementById('sample_urine');
    const sampleSwabCheckbox = document.getElementById('sample_swab');

    if (sampleBloodCheckbox) {
        sampleBloodCheckbox.addEventListener('change', function() {
            toggleSampleTypeOptions('blood', this.checked);
        });
    }

    if (sampleUrineCheckbox) {
        sampleUrineCheckbox.addEventListener('change', function() {
            toggleSampleTypeOptions('urine', this.checked);
        });
    }

    if (sampleSwabCheckbox) {
        sampleSwabCheckbox.addEventListener('change', function() {
            toggleSampleTypeOptions('swab', this.checked);
        });
    }

    // Initialize existing selections
    if (sampleBloodCheckbox && sampleBloodCheckbox.checked) {
        toggleSampleTypeOptions('blood', true);
    }
    if (sampleUrineCheckbox && sampleUrineCheckbox.checked) {
        toggleSampleTypeOptions('urine', true);
    }
    if (sampleSwabCheckbox && sampleSwabCheckbox.checked) {
        toggleSampleTypeOptions('swab', true);
    }
}

function toggleSampleTypeOptions(sampleType, show) {
    const containerId = `${sampleType}_options_container`;
    let container = document.getElementById(containerId);

    if (show) {
        if (!container) {
            container = createSampleTypeContainer(sampleType);
        }
        container.style.display = 'block';
        loadSampleTypeOptions(sampleType);
    } else {
        if (container) {
            container.style.display = 'none';
        }
    }
}

function createSampleTypeContainer(sampleType) {
    const sampleCheckboxesDiv = document.querySelector('.sample-checkboxes');
    if (!sampleCheckboxesDiv) return null;

    const container = document.createElement('div');
    container.id = `${sampleType}_options_container`;
    container.className = 'sample-type-options mt-3';
    container.style.display = 'none';
    container.style.border = '1px solid #dee2e6';
    container.style.borderRadius = '4px';
    container.style.padding = '15px';
    container.style.backgroundColor = '#f8f9fa';

    const title = document.createElement('h6');
    title.className = 'mb-3';
    title.style.color = '#495057';
    title.style.fontWeight = '600';
    title.textContent = getSampleTypeTitle(sampleType);

    const optionsDiv = document.createElement('div');
    optionsDiv.id = `${sampleType}_options`;
    optionsDiv.className = 'sample-options';

    container.appendChild(title);
    container.appendChild(optionsDiv);

    // Insert after the sample checkboxes
    sampleCheckboxesDiv.parentNode.insertBefore(container, sampleCheckboxesDiv.nextSibling);

    return container;
}

function getSampleTypeTitle(sampleType) {
    const titles = {
        'swab': 'Select Swab Type:',
        'urine': 'Select Urine Test Type:',
        'blood': 'Select Blood Test Type:'
    };
    return titles[sampleType] || 'Select Options:';
}

// function loadSampleTypeOptions(sampleType) {
//     const optionsDiv = document.getElementById(`${sampleType}_options`);
//     if (!optionsDiv) return;

//     // Show loading
//     optionsDiv.innerHTML = '<div class="text-muted">Loading options...</div>';

//     // Fetch options from API
//     fetch(`/api/lab/${sampleType}-types`)
//         .then(response => response.json())
//         .then(data => {
//             if (data.success) {
//                 renderSampleTypeOptions(sampleType, data.data);
//             } else {
//                 optionsDiv.innerHTML = '<div class="text-danger">Failed to load options</div>';
//             }
//         })
//         .catch(error => {
//             console.error('Error loading sample type options:', error);
//             optionsDiv.innerHTML = '<div class="text-danger">Error loading options</div>';
//         });
// }

async function loadSampleTypeOptions(sampleType) {
  const optionsDiv = document.getElementById(`${sampleType}_options`);
  if (!optionsDiv) return;

  optionsDiv.innerHTML = '<div class="text-muted">Loading options...</div>';

  // Hard-coded sample type options
  const sampleTypeData = {
    blood: [
      { ID: 1, Name: "Complete Blood Count (CBC)", Category: { String: "Hematology" }, Description: { String: "Measures red blood cells, white blood cells, and platelets" } },
      { ID: 2, Name: "Hemoglobin", Category: { String: "Hematology" }, Description: { String: "Measures oxygen-carrying protein in red blood cells" } },
      { ID: 3, Name: "Hematocrit", Category: { String: "Hematology" }, Description: { String: "Percentage of blood volume occupied by red blood cells" } },
      { ID: 4, Name: "White Blood Cell Count", Category: { String: "Hematology" }, Description: { String: "Count of white blood cells in the blood" } },
      { ID: 5, Name: "Platelet Count", Category: { String: "Hematology" }, Description: { String: "Count of platelets in the blood" } },
      { ID: 6, Name: "Blood Glucose", Category: { String: "Chemistry" }, Description: { String: "Measures sugar levels in the blood" } },
      { ID: 7, Name: "Creatinine", Category: { String: "Chemistry" }, Description: { String: "Measures kidney function" } },
      { ID: 8, Name: "Urea", Category: { String: "Chemistry" }, Description: { String: "Measures kidney function and protein metabolism" } },
      { ID: 9, Name: "Sodium", Category: { String: "Electrolytes" }, Description: { String: "Measures sodium levels in the blood" } },
      { ID: 10, Name: "Potassium", Category: { String: "Electrolytes" }, Description: { String: "Measures potassium levels in the blood" } },
      { ID: 11, Name: "Chloride", Category: { String: "Electrolytes" }, Description: { String: "Measures chloride levels in the blood" } },
      { ID: 12, Name: "Bicarbonate", Category: { String: "Electrolytes" }, Description: { String: "Measures bicarbonate levels in the blood" } },
      { ID: 13, Name: "Liver Function Tests", Category: { String: "Liver" }, Description: { String: "ALT, AST, ALP, Bilirubin, Albumin" } },
      { ID: 14, Name: "C-Reactive Protein", Category: { String: "Inflammation" }, Description: { String: "Measures inflammation in the body" } },
      { ID: 15, Name: "Erythrocyte Sedimentation Rate", Category: { String: "Inflammation" }, Description: { String: "Measures inflammation and infection" } }
    ],
    swab: [
      { ID: 1, Name: "Nasopharyngeal Swab", Description: { String: "Swab from the back of the nose and throat" } },
      { ID: 2, Name: "Oropharyngeal Swab", Description: { String: "Swab from the back of the throat" } },
      { ID: 3, Name: "Buccal Swab", Description: { String: "Swab from the inside of the cheek" } },
      { ID: 4, Name: "Nasal Swab", Description: { String: "Swab from the nostrils" } },
      { ID: 5, Name: "Throat Swab", Description: { String: "Swab from the throat area" } }
    ],
    urine: [
      { ID: 1, Name: "Urinalysis", Description: { String: "Complete urine analysis including pH, specific gravity, protein, glucose, ketones, blood, leukocytes, nitrites" } },
      { ID: 2, Name: "Urine Culture", Description: { String: "Bacterial culture and sensitivity testing" } },
      { ID: 3, Name: "Urine Microscopy", Description: { String: "Microscopic examination of urine sediment" } },
      { ID: 4, Name: "Urine Protein", Description: { String: "Measurement of protein in urine" } },
      { ID: 5, Name: "Urine Glucose", Description: { String: "Measurement of glucose in urine" } },
      { ID: 6, Name: "Urine Ketones", Description: { String: "Measurement of ketones in urine" } }
    ]
  };

  try {
    const data = sampleTypeData[sampleType];
    if (data) {
      renderSampleTypeOptions(sampleType, data);
    } else {
      throw new Error(`Unknown sample type: ${sampleType}`);
    }
  } catch (err) {
    console.error('Error loading sample type options:', err);
    optionsDiv.innerHTML = `<div class="text-danger">Failed to load options: ${err.message}</div>`;
  }
}
  

function renderSampleTypeOptions(sampleType, options) {
    const optionsDiv = document.getElementById(`${sampleType}_options`);
    if (!optionsDiv) return;

    let html = '';

    if (sampleType === 'blood') {
        // Group blood tests by category
        const categories = groupBloodTestsByCategory(options);
        html = renderBloodTestOptions(categories);
    } else {
        // Simple checkbox list for swab and urine
        html = renderSimpleOptions(sampleType, options);
    }

    optionsDiv.innerHTML = html;
    attachOptionEventListeners(sampleType);
}

function groupBloodTestsByCategory(bloodTests) {
    const categories = {};
    bloodTests.forEach(test => {
        const category = test.Category.String || 'Other';
        if (!categories[category]) {
            categories[category] = [];
        }
        categories[category].push(test);
    });
    return categories;
}

function renderBloodTestOptions(categories) {
    let html = '';
    
    Object.keys(categories).forEach(category => {
        html += `<div class="blood-category mb-3">`;
        html += `<h6 class="category-title" style="color: #6c757d; font-size: 0.9rem; margin-bottom: 8px;">${category}:</h6>`;
        html += `<div class="category-options">`;
        
        categories[category].forEach(test => {
            html += `<div class="form-check">`;
            html += `<input class="form-check-input blood-test-checkbox" type="checkbox" value="${test.ID}" id="blood_${test.ID}" name="blood_tests[]">`;
            html += `<label class="form-check-label" for="blood_${test.ID}" style="font-size: 0.85rem;">`;
            html += `${test.Name}`;
            if (test.Description.String) {
                html += `<small class="text-muted d-block">${test.Description.String}</small>`;
            }
            html += `</label>`;
            html += `</div>`;
        });
        
        html += `</div></div>`;
    });

    // Add "Others (Specify)" option
    html += `<div class="form-check">`;
    html += `<input class="form-check-input" type="checkbox" id="blood_other" name="blood_tests[]" value="other">`;
    html += `<label class="form-check-label" for="blood_other" style="font-size: 0.85rem;">Others (Specify)</label>`;
    html += `</div>`;
    html += `<div class="mt-2" id="blood_other_specify_container" style="display: none;">`;
    html += `<input type="text" class="form-control form-control-sm" id="blood_other_specify" placeholder="Specify other blood tests" style="font-size: 0.85rem;">`;
    html += `</div>`;

    return html;
}

function renderSimpleOptions(sampleType, options) {
    let html = '';
    
    options.forEach(option => {
        html += `<div class="form-check">`;
        html += `<input class="form-check-input ${sampleType}-test-checkbox" type="checkbox" value="${option.ID}" id="${sampleType}_${option.ID}" name="${sampleType}_tests[]">`;
        html += `<label class="form-check-label" for="${sampleType}_${option.ID}" style="font-size: 0.85rem;">`;
        html += `${option.Name}`;
        if (option.Description && option.Description.String) {
            html += `<small class="text-muted d-block">${option.Description.String}</small>`;
        }
        html += `</label>`;
        html += `</div>`;
    });

    // Add "Others (Specify)" option
    html += `<div class="form-check">`;
    html += `<input class="form-check-input" type="checkbox" id="${sampleType}_other" name="${sampleType}_tests[]" value="other">`;
    html += `<label class="form-check-label" for="${sampleType}_other" style="font-size: 0.85rem;">Others (Specify)</label>`;
    html += `</div>`;
    html += `<div class="mt-2" id="${sampleType}_other_specify_container" style="display: none;">`;
    html += `<input type="text" class="form-control form-control-sm" id="${sampleType}_other_specify" placeholder="Specify other ${sampleType} tests" style="font-size: 0.85rem;">`;
    html += `</div>`;

    return html;
}

function attachOptionEventListeners(sampleType) {
    // Handle "Others (Specify)" checkbox
    const otherCheckbox = document.getElementById(`${sampleType}_other`);
    const otherSpecifyContainer = document.getElementById(`${sampleType}_other_specify_container`);
    
    if (otherCheckbox && otherSpecifyContainer) {
        otherCheckbox.addEventListener('change', function() {
            otherSpecifyContainer.style.display = this.checked ? 'block' : 'none';
        });
    }

    // Handle test checkboxes
    const checkboxes = document.querySelectorAll(`.${sampleType}-test-checkbox, .blood-test-checkbox`);
    checkboxes.forEach(checkbox => {
        checkbox.addEventListener('change', function() {
            saveSampleSelections(sampleType);
        });
    });
}

function saveSampleSelections(sampleType) {
    const checkboxes = document.querySelectorAll(`input[name="${sampleType}_tests[]"]:checked`);
    const selections = [];

    checkboxes.forEach(checkbox => {
        if (checkbox.value === 'other') {
            const otherSpecify = document.getElementById(`${sampleType}_other_specify`);
            if (otherSpecify && otherSpecify.value.trim()) {
                selections.push({
                    selected_type_id: 0, // 0 for "Others"
                    other_specify: otherSpecify.value.trim()
                });
            }
        } else {
            selections.push({
                selected_type_id: parseInt(checkbox.value),
                other_specify: ''
            });
        }
    });

    // Get lab_id from the form
    const labIdInput = document.querySelector('input[name="lab_id"]');
    const labId = labIdInput ? parseInt(labIdInput.value) : 0;

    if (labId > 0 && selections.length > 0) {
        // Save selections to database
        fetch('/api/lab/sample-selections', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                lab_id: labId,
                sample_type: sampleType,
                selections: selections
            })
        })
        .then(response => response.json())
        .then(data => {
            if (!data.success) {
                console.error('Failed to save sample selections:', data.error);
            }
        })
        .catch(error => {
            console.error('Error saving sample selections:', error);
        });
    }
}

// Function to load existing selections when editing
function loadExistingSelections(labId) {
    if (!labId) return;

    fetch(`/api/lab/sample-selections/${labId}`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                data.data.forEach(selection => {
                    // Check the appropriate sample type checkbox
                    const sampleCheckbox = document.getElementById(`sample_${selection.sample_type}`);
                    if (sampleCheckbox) {
                        sampleCheckbox.checked = true;
                        toggleSampleTypeOptions(selection.sample_type, true);
                        
                        // Wait for options to load, then check the selected ones
                        setTimeout(() => {
                            if (selection.selected_type_id > 0) {
                                const optionCheckbox = document.getElementById(`${selection.sample_type}_${selection.selected_type_id}`);
                                if (optionCheckbox) {
                                    optionCheckbox.checked = true;
                                }
                            } else if (selection.other_specify) {
                                const otherCheckbox = document.getElementById(`${selection.sample_type}_other`);
                                const otherSpecify = document.getElementById(`${selection.sample_type}_other_specify`);
                                if (otherCheckbox && otherSpecify) {
                                    otherCheckbox.checked = true;
                                    otherSpecify.value = selection.other_specify;
                                    otherSpecify.parentElement.style.display = 'block';
                                }
                            }
                        }, 500);
                    }
                });
            }
        })
        .catch(error => {
            console.error('Error loading existing selections:', error);
        });
} 