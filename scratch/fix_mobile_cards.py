import os

css_to_inject = """
  /* --- Mobile Card Layout Injected --- */
  @media (max-width: 768px) {
    .table-container {
      background: transparent !important;
      border: none !important;
      box-shadow: none !important;
      padding: 0 !important;
    }
    .data-table, .data-table tbody, .data-table tr, .data-table td {
      display: block;
      width: 100%;
    }
    .data-table thead {
      display: none;
    }
    .data-table tr {
      margin-bottom: 1.25rem;
      border: 1px solid #cbd5e1;
      border-radius: 12px;
      background: #fff;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
      overflow: hidden;
    }
    .data-table td {
      padding: 1rem;
      border-bottom: 1px solid #f1f5f9;
      text-align: left !important;
    }
    .data-table td:last-child {
      border-bottom: none;
      background-color: #f8fafc;
    }
    .siswa-nama {
      font-size: 1.1rem;
      color: #0f172a;
    }
    .page-header-actions {
      flex-direction: column;
      align-items: stretch;
    }
    .action-buttons {
      width: 100%;
      flex-direction: column;
      margin-top: 1rem;
    }
    .btn-action-primary, .btn-action-secondary {
      width: 100%;
      justify-content: center;
    }
    .search-bar {
      width: 100%;
    }
    textarea, input[type="number"], .form-input {
      width: 100% !important;
      min-width: 0 !important;
    }
    .student-row {
      flex-direction: column;
      align-items: stretch;
    }
    .student-info-group {
      margin-bottom: 0.5rem;
    }
  }
</style>
"""

def process_directory(directory):
    for root, _, files in os.walk(directory):
        for file in files:
            if file == "+page.svelte":
                file_path = os.path.join(root, file)
                with open(file_path, "r") as f:
                    content = f.read()
                
                # Check if already injected to avoid duplicates
                if "/* --- Mobile Card Layout Injected --- */" in content:
                    print(f"Skipping (already injected): {file_path}")
                    continue
                
                # We need to find the last </style>
                if "</style>" in content:
                    # Replace the last occurrence of </style>
                    last_index = content.rfind("</style>")
                    new_content = content[:last_index] + css_to_inject + content[last_index + 8:]
                    
                    with open(file_path, "w") as f:
                        f.write(new_content)
                    print(f"Injected into: {file_path}")

base_dirs = [
    "frontend/src/routes/guru/input-nilai",
    "frontend/src/routes/guru/rapor"
]

for d in base_dirs:
    process_directory(d)
