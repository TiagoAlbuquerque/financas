import './style.css';
import './app.css';
import { Chart, registerables } from 'chart.js';

// Register Chart.js components
Chart.register(...registerables);

// Import Wails backend functions
import {
  GetOverallSummary,
  GetYearSummary,
  GetMonthSummary,
  GetMonthTransactions,
  SaveTransaction,
  DeleteTransaction,
  GetDataDir
} from '../wailsjs/go/main/App';

// Global state
const state = {
  currentView: 'overall', // 'overall', 'year', 'month'
  selectedYear: null,
  selectedMonth: null,
  overallSummary: null,
  yearSummary: null,
  monthSummary: null,
  monthTransactions: [],
  charts: {
    main: null,
    category: null
  }
};

// Portuguese month names
const MONTH_NAMES = {
  "01": "Janeiro",
  "02": "Fevereiro",
  "03": "Março",
  "04": "Abril",
  "05": "Maio",
  "06": "Junho",
  "07": "Julho",
  "08": "Agosto",
  "09": "Setembro",
  "10": "Outubro",
  "11": "Novembro",
  "12": "Dezembro"
};

// Start application
document.addEventListener('DOMContentLoaded', async () => {
  // Load data path
  try {
    const path = await GetDataDir();
    document.getElementById('storage-path').innerText = path;
    document.getElementById('storage-path').title = path;
  } catch (err) {
    console.error("Erro ao carregar caminho de armazenamento:", err);
  }

  // Setup DOM Event Listeners
  setupEventListeners();

  // Load initial view
  showOverallView();
});

function setupEventListeners() {
  // Navigation sidebar
  document.getElementById('btn-nav-overall').addEventListener('click', () => {
    showOverallView();
  });

  // Reload button
  document.getElementById('btn-reload').addEventListener('click', async () => {
    const btn = document.getElementById('btn-reload');
    const originalHTML = btn.innerHTML;
    
    btn.innerHTML = '<span class="icon" style="display:inline-block; animation: spin 0.8s linear infinite;">🔄</span> Atualizando...';
    btn.disabled = true;
    
    await refreshData();
    
    // Smooth transition back
    setTimeout(() => {
      btn.innerHTML = originalHTML;
      btn.disabled = false;
    }, 450);
  });

  // Modal open/close
  const modal = document.getElementById('transaction-modal');
  document.getElementById('btn-add-transaction').addEventListener('click', () => {
    // Set default date to today in YYYY-MM-DD
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('t-date').value = today;
    
    // Clear hidden inputs for editing
    document.getElementById('t-id').value = '';
    document.getElementById('t-orig-year').value = '';
    document.getElementById('t-orig-month').value = '';
    document.getElementById('modal-title').innerText = "Adicionar Transação";
    
    // Clear inputs
    document.getElementById('t-category').value = '';
    document.getElementById('t-description').value = '';
    document.getElementById('t-amount').value = '';
    
    modal.classList.add('open');
  });

  const closeModal = () => modal.classList.remove('open');
  document.getElementById('modal-close').addEventListener('click', closeModal);
  document.getElementById('btn-cancel-modal').addEventListener('click', closeModal);
  
  // Close modal when clicking outside card
  modal.addEventListener('click', (e) => {
    if (e.target === modal) closeModal();
  });

  // Form submission
  document.getElementById('transaction-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const id = document.getElementById('t-id').value;
    const origYear = parseInt(document.getElementById('t-orig-year').value);
    const origMonth = parseInt(document.getElementById('t-orig-month').value);
    
    const date = document.getElementById('t-date').value;
    const type = document.querySelector('input[name="t-type"]:checked').value;
    const category = document.getElementById('t-category').value.trim();
    const description = document.getElementById('t-description').value.trim();
    const amount = parseFloat(document.getElementById('t-amount').value);

    if (!date || !category || !description || isNaN(amount)) {
      alert("Por favor, preencha todos os campos corretamente.");
      return;
    }

    try {
      // If we are editing and the date was changed to a different month/year, 
      // we must delete the transaction from the old location first to avoid duplicates
      if (id && !isNaN(origYear) && !isNaN(origMonth)) {
        const [newY, newM] = date.split('-').map(Number);
        if (newY !== origYear || newM !== origMonth) {
          await DeleteTransaction(id, origYear, origMonth);
        }
      }

      await SaveTransaction({
        id: id || "",
        date,
        type,
        category,
        description,
        amount
      });
      
      closeModal();
      
      // Reload current view
      await refreshData();
    } catch (err) {
      alert("Erro ao salvar transação: " + err);
    }
  });

  // Filters
  document.getElementById('filter-search').addEventListener('input', () => {
    renderTransactionsTable();
  });
  document.getElementById('filter-category').addEventListener('change', () => {
    renderTransactionsTable();
  });
}

// State navigation actions
async function showOverallView() {
  state.currentView = 'overall';
  state.selectedYear = null;
  state.selectedMonth = null;
  
  // Highlight navigation item
  document.getElementById('btn-nav-overall').classList.add('active');
  
  await refreshData();
}

async function showYearView(year) {
  state.currentView = 'year';
  state.selectedYear = year;
  state.selectedMonth = null;

  // Un-highlight overall nav item
  document.getElementById('btn-nav-overall').classList.remove('active');

  await refreshData();
}

async function showMonthView(year, month) {
  state.currentView = 'month';
  state.selectedYear = year;
  state.selectedMonth = parseInt(month);

  // Un-highlight overall nav item
  document.getElementById('btn-nav-overall').classList.remove('active');

  await refreshData();
}

// Data loaders
async function refreshData() {
  try {
    // 1. Always load overall summary to keep sidebar folder structure updated
    state.overallSummary = await GetOverallSummary();
    
    // 2. Load view-specific data
    if (state.currentView === 'year') {
      state.yearSummary = await GetYearSummary(state.selectedYear);
    } else if (state.currentView === 'month') {
      state.monthSummary = await GetMonthSummary(state.selectedYear, state.selectedMonth);
      const details = await GetMonthTransactions(state.selectedYear, state.selectedMonth);
      state.monthTransactions = details.transactions || [];
    }
    
    // 3. Render everything
    updateSidebar();
    updateDashboardUI();
  } catch (err) {
    console.error("Erro ao atualizar dados:", err);
  }
}

// Render folder structure in sidebar
function updateSidebar() {
  const listContainer = document.getElementById('years-list');
  listContainer.innerHTML = '';

  const years = Object.keys(state.overallSummary.years || {}).sort((a, b) => b - a);
  
  if (years.length === 0) {
    listContainer.innerHTML = '<p style="font-size:0.8rem; color:var(--text-secondary); padding:8px 10px;">Nenhum dado cadastrado</p>';
    return;
  }

  years.forEach(year => {
    const yearInt = parseInt(year);
    const isYearActive = state.currentView !== 'overall' && state.selectedYear === yearInt;
    
    const yearGroup = document.createElement('div');
    yearGroup.className = `year-group ${isYearActive ? 'open' : ''}`;
    
    // Year header button
    const header = document.createElement('button');
    header.className = 'year-header';
    header.innerHTML = `
      <span>📅 ${year}</span>
      <span class="arrow">▶</span>
    `;
    
    // Click on year header expands/collapses and opens Year view
    header.addEventListener('click', (e) => {
      // Toggle collapse class locally
      yearGroup.classList.toggle('open');
      showYearView(yearInt);
    });
    
    const monthsList = document.createElement('div');
    monthsList.className = 'months-sublist';
    
    // For months, Wails backend calculates what months actually have summaries/details.
    // However, let's show months 01-12 or just the months that have data.
    // Showing months with data is cleaner, but let's show all 12 months for active year to let users browse/add transactions easily!
    for (let m = 1; m <= 12; m++) {
      const mStr = m.toString().padStart(2, '0');
      const monthItem = document.createElement('button');
      const isMonthActive = state.currentView === 'month' && state.selectedYear === yearInt && state.selectedMonth === m;
      
      monthItem.className = `month-item ${isMonthActive ? 'active' : ''}`;
      monthItem.innerText = MONTH_NAMES[mStr];
      
      monthItem.addEventListener('click', (e) => {
        e.stopPropagation(); // Avoid triggering year collapse
        showMonthView(yearInt, mStr);
      });
      
      monthsList.appendChild(monthItem);
    }
    
    yearGroup.appendChild(header);
    yearGroup.appendChild(monthsList);
    listContainer.appendChild(yearGroup);
  });
}

// Refresh Dashboard UI elements (Title, Cards, Charts, Tables)
function updateDashboardUI() {
  const titleEl = document.getElementById('dashboard-title');
  const subtitleEl = document.getElementById('dashboard-subtitle');
  
  const cardIncome = document.getElementById('card-total-income');
  const cardExpense = document.getElementById('card-total-expense');
  const cardBalance = document.getElementById('card-total-balance');
  const cardBalanceMeta = document.getElementById('card-balance-meta');
  
  const mainChartBox = document.querySelector('.main-chart-box');
  const pieChartBox = document.querySelector('.pie-chart-box');
  const detailsSection = document.querySelector('.details-section');

  // Format currency helpers
  const formatCurrency = (val) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val || 0);

  let income = 0;
  let expense = 0;
  let balance = 0;

  if (state.currentView === 'overall') {
    titleEl.innerText = "Visão Geral";
    subtitleEl.innerText = "Painel acumulado de todas as finanças registradas";
    
    income = state.overallSummary.earnings;
    expense = state.overallSummary.expenses;
    
    // Show Main Chart (Yearly comparison), Hide Category chart and Transaction table
    mainChartBox.style.display = 'block';
    pieChartBox.style.display = 'none';
    detailsSection.style.display = 'none';
    
    document.getElementById('main-chart-title').innerText = "Comparação Multianual";
    renderOverallChart();
    
  } else if (state.currentView === 'year') {
    titleEl.innerText = `Ano ${state.selectedYear}`;
    subtitleEl.innerText = `Resumo de desempenho para o ano de ${state.selectedYear}`;
    
    income = state.yearSummary.earnings;
    expense = state.yearSummary.expenses;
    
    // Show Main Chart (Monthly comparison) and Show Category chart (Doughnut)
    mainChartBox.style.display = 'block';
    pieChartBox.style.display = 'block';
    detailsSection.style.display = 'none';
    
    document.getElementById('main-chart-title').innerText = `Desempenho de ${state.selectedYear}`;
    renderYearChart();
    renderCategoryChart();
    
  } else if (state.currentView === 'month') {
    const mStr = state.selectedMonth.toString().padStart(2, '0');
    titleEl.innerText = `${MONTH_NAMES[mStr]} de ${state.selectedYear}`;
    subtitleEl.innerText = `Detalhamento de receitas, despesas e transações individuais`;
    
    income = state.monthSummary.earnings;
    expense = state.monthSummary.expenses;
    
    // In Monthly view, hide the main monthly comparison bar chart.
    // Instead, show category breakdown chart and the full transaction details table.
    mainChartBox.style.display = 'none';
    pieChartBox.style.display = 'block';
    detailsSection.style.display = 'block';
    
    // Refresh filter category options
    populateCategoryFilters();
    renderTransactionsTable();
    renderCategoryChart();
  }

  balance = income - expense;

  cardIncome.innerText = formatCurrency(income);
  cardExpense.innerText = formatCurrency(expense);
  cardBalance.innerText = formatCurrency(balance);

  // Style Balance Card based on net total
  const balanceCard = document.querySelector('.card-balance');
  balanceCard.classList.remove('card-balance-positive', 'card-balance-negative');
  if (balance >= 0) {
    cardBalanceMeta.innerText = "Saldo positivo acumulado";
    cardBalanceMeta.style.color = "var(--color-income)";
    cardBalance.style.color = "var(--color-income)";
  } else {
    cardBalanceMeta.innerText = "Saldo negativo (Atenção)";
    cardBalanceMeta.style.color = "var(--color-expense)";
    cardBalance.style.color = "var(--color-expense)";
  }
}

// Chart 1: Compare earnings/expenses across all years (Overall View)
function renderOverallChart() {
  const ctx = document.getElementById('mainChart').getContext('2d');
  
  if (state.charts.main) {
    state.charts.main.destroy();
  }

  const yearsData = state.overallSummary.years || {};
  const labels = Object.keys(yearsData).sort();
  const incomes = labels.map(y => yearsData[y].earnings);
  const expenses = labels.map(y => yearsData[y].expenses);

  state.charts.main = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [
        {
          label: 'Receitas',
          data: incomes,
          backgroundColor: 'rgba(16, 185, 129, 0.75)',
          borderColor: '#10b981',
          borderWidth: 1,
          borderRadius: 4
        },
        {
          label: 'Despesas',
          data: expenses,
          backgroundColor: 'rgba(244, 63, 94, 0.75)',
          borderColor: '#f43f5e',
          borderWidth: 1,
          borderRadius: 4
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          labels: { color: '#9ca3af', font: { family: 'Nunito' } }
        }
      },
      scales: {
        x: {
          grid: { display: false },
          ticks: { color: '#9ca3af', font: { family: 'Nunito' } }
        },
        y: {
          grid: { color: 'rgba(255, 255, 255, 0.05)' },
          ticks: { color: '#9ca3af', font: { family: 'Nunito' } }
        }
      }
    }
  });
}

// Chart 2: Compare earnings/expenses monthly (Yearly View)
function renderYearChart() {
  const ctx = document.getElementById('mainChart').getContext('2d');
  
  if (state.charts.main) {
    state.charts.main.destroy();
  }

  const monthsData = state.yearSummary.months || {};
  const labels = Object.keys(MONTH_NAMES);
  const labelNames = labels.map(m => MONTH_NAMES[m].substring(0, 3)); // Jan, Fev...
  
  const incomes = labels.map(m => monthsData[m] ? monthsData[m].earnings : 0);
  const expenses = labels.map(m => monthsData[m] ? monthsData[m].expenses : 0);

  state.charts.main = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labelNames,
      datasets: [
        {
          label: 'Receitas',
          data: incomes,
          backgroundColor: 'rgba(16, 185, 129, 0.75)',
          borderColor: '#10b981',
          borderWidth: 1,
          borderRadius: 4
        },
        {
          label: 'Despesas',
          data: expenses,
          backgroundColor: 'rgba(244, 63, 94, 0.75)',
          borderColor: '#f43f5e',
          borderWidth: 1,
          borderRadius: 4
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          labels: { color: '#9ca3af', font: { family: 'Nunito' } }
        }
      },
      scales: {
        x: {
          grid: { display: false },
          ticks: { color: '#9ca3af', font: { family: 'Nunito' } }
        },
        y: {
          grid: { color: 'rgba(255, 255, 255, 0.05)' },
          ticks: { color: '#9ca3af', font: { family: 'Nunito' } }
        }
      }
    }
  });
}

// Chart 3: Doughnut Chart of expenses by category (Monthly View)
function renderCategoryChart() {
  const ctx = document.getElementById('categoryChart').getContext('2d');
  
  if (state.charts.category) {
    state.charts.category.destroy();
  }

  let expensesData = {};
  if (state.currentView === 'month') {
    expensesData = state.monthSummary.expenses_by_category || {};
  } else if (state.currentView === 'year') {
    expensesData = state.yearSummary.expenses_by_category || {};
  }
  const categories = Object.keys(expensesData);
  const values = categories.map(cat => expensesData[cat]);

  if (categories.length === 0) {
    // Empty state
    state.charts.category = new Chart(ctx, {
      type: 'doughnut',
      data: {
        labels: ['Sem despesas'],
        datasets: [{
          data: [1],
          backgroundColor: ['rgba(255, 255, 255, 0.1)'],
          borderWidth: 0
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { labels: { color: '#9ca3af' } }
        }
      }
    });
    return;
  }

  // Generate pretty gradient colors
  const colorPalette = [
    '#f43f5e', '#3b82f6', '#ec4899', '#eab308', '#a855f7',
    '#f97316', '#06b6d4', '#84cc16', '#14b8a6', '#64748b'
  ];

  state.charts.category = new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: categories,
      datasets: [{
        data: values,
        backgroundColor: colorPalette.slice(0, categories.length),
        borderColor: 'rgba(19, 27, 46, 0.8)',
        borderWidth: 2
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'right',
          labels: {
            color: '#9ca3af',
            font: { family: 'Nunito', size: 11 }
          }
        }
      }
    }
  });
}

// Populate search drop-down filter with unique categories of the selected month
function populateCategoryFilters() {
  const select = document.getElementById('filter-category');
  const currentValue = select.value;
  select.innerHTML = '<option value="">Todas as Categorias</option>';

  const categories = new Set();
  state.monthTransactions.forEach(t => {
    if (t.category) categories.add(t.category);
  });

  Array.from(categories).sort().forEach(cat => {
    const opt = document.createElement('option');
    opt.value = cat;
    opt.innerText = cat;
    select.appendChild(opt);
  });

  // Restore previous filter selection if still valid
  if (categories.has(currentValue)) {
    select.value = currentValue;
  }
}

// Render Transactions list inside monthly view
function renderTransactionsTable() {
  const tbody = document.getElementById('transactions-body');
  const emptyMessage = document.getElementById('no-transactions-message');
  
  tbody.innerHTML = '';
  
  const searchVal = document.getElementById('filter-search').value.toLowerCase().trim();
  const catVal = document.getElementById('filter-category').value;

  // Filter transactions in memory
  const filtered = state.monthTransactions.filter(t => {
    const matchesSearch = t.description.toLowerCase().includes(searchVal);
    const matchesCategory = catVal === "" || t.category === catVal;
    return matchesSearch && matchesCategory;
  });

  if (filtered.length === 0) {
    emptyMessage.style.display = 'block';
    return;
  }
  
  emptyMessage.style.display = 'none';

  // Sort by date (descending)
  filtered.sort((a, b) => b.date.localeCompare(a.date));

  const formatCurrency = (val) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val);
  const formatDate = (dateStr) => {
    const [y, m, d] = dateStr.split('-');
    return `${d}/${m}/${y}`;
  };

  filtered.forEach(t => {
    const tr = document.createElement('tr');
    
    const isEarning = t.type === 'earning';
    const typeText = isEarning ? "Receita" : "Despesa";
    const badgeClass = isEarning ? "badge-income" : "badge-expense";
    const amountClass = isEarning ? "amount-income" : "amount-expense";
    const prefixSign = isEarning ? "+" : "-";

    tr.innerHTML = `
      <td>${formatDate(t.date)}</td>
      <td><strong>${escapeHTML(t.description)}</strong></td>
      <td><span style="color:var(--text-secondary);">${escapeHTML(t.category)}</span></td>
      <td><span class="badge ${badgeClass}">${typeText}</span></td>
      <td class="text-right ${amountClass}">${prefixSign} ${formatCurrency(t.amount)}</td>
      <td class="text-center">
        <button class="btn-edit-icon" title="Editar Transação">
          ✏️
        </button>
        <button class="btn-danger-icon" title="Excluir Transação">
          🗑️
        </button>
      </td>
    `;

    // Bind edit action
    tr.querySelector('.btn-edit-icon').addEventListener('click', () => {
      document.getElementById('modal-title').innerText = "Editar Transação";
      document.getElementById('t-id').value = t.id;
      document.getElementById('t-orig-year').value = state.selectedYear;
      document.getElementById('t-orig-month').value = state.selectedMonth;
      document.getElementById('t-date').value = t.date;
      
      // Set type radio
      document.querySelector(`input[name="t-type"][value="${t.type}"]`).checked = true;
      
      document.getElementById('t-category').value = t.category;
      document.getElementById('t-description').value = t.description;
      document.getElementById('t-amount').value = t.amount;
      
      document.getElementById('transaction-modal').classList.add('open');
    });

    // Bind delete action
    tr.querySelector('.btn-danger-icon').addEventListener('click', async () => {
      if (confirm(`Deseja excluir a transação "${t.description}" de ${formatCurrency(t.amount)}?`)) {
        try {
          await DeleteTransaction(t.id, state.selectedYear, state.selectedMonth);
          await refreshData();
        } catch (err) {
          alert("Erro ao excluir transação: " + err);
        }
      }
    });

    tbody.appendChild(tr);
  });
}

// Simple HTML escaping helper
function escapeHTML(str) {
  return str.replace(/[&<>'"]/g, 
    tag => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      "'": '&#39;',
      '"': '&quot;'
    }[tag] || tag)
  );
}
