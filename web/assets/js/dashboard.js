// Dashboard charts — lê dados de window.dashboardData definido inline no HTML
(function () {
  if (typeof ApexCharts === 'undefined') {
    console.error('ApexCharts não carregado');
    return;
  }

  const data = window.dashboardData || {};
  const dailyPurchasesData = data.dailyPurchases || [];
  const dailyDownloadsData = data.dailyDownloads || [];
  const topEbooksData = data.topEbooks || [];
  const topDownloadedEbooksData = data.topDownloadedEbooks || [];

  // Últimos 7 dias
  function getLast7Days() {
    const dates = [];
    const displayDates = [];
    for (let i = 6; i >= 0; i--) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      dates.push(date.toISOString().split('T')[0]);
      const day = String(date.getDate()).padStart(2, '0');
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const year = date.getFullYear();
      displayDates.push(`${day}/${month}/${year}`);
    }
    return { dates, displayDates };
  }

  const { dates: last7Days, displayDates: last7DaysDisplay } = getLast7Days();

  const purchasesMap = new Map(dailyPurchasesData.map(item => [item.date, item.count]));
  const downloadsMap = new Map(dailyDownloadsData.map(item => [item.date, item.count]));

  const dailyPurchasesCounts = last7Days.map(date => purchasesMap.get(date) || 0);
  const dailyDownloadsCounts = last7Days.map(date => downloadsMap.get(date) || 0);

  const topEbooksLabels = topEbooksData.map(item => item.title);
  const topEbooksCounts = topEbooksData.map(item => item.total_purchases);

  const topDownloadedEbooksLabels = topDownloadedEbooksData.map(item => item.title);
  const topDownloadedEbooksCounts = topDownloadedEbooksData.map(item => item.total_downloads);

  const intFormatter = val => Number.isInteger(val) ? val : '';

  const areaDefaults = {
    chart: { type: 'area', height: 300, toolbar: { show: false } },
    fill: {
      type: 'gradient',
      gradient: { shadeIntensity: 1, opacityFrom: 0.7, opacityTo: 0.2, stops: [0, 90, 100] }
    },
    dataLabels: { enabled: false },
    stroke: { curve: 'smooth', width: 3 },
    xaxis: {
      categories: last7DaysDisplay,
      labels: { style: { colors: '#6B7280', fontSize: '12px' } }
    },
    yaxis: {
      labels: { style: { colors: '#6B7280', fontSize: '12px' }, formatter: intFormatter },
      min: 0,
      forceNiceScale: true
    },
    grid: { borderColor: '#E5E7EB', strokeDashArray: 4 },
    tooltip: { theme: 'light', x: { show: false } }
  };

  const barDefaults = {
    chart: { type: 'bar', height: 300, toolbar: { show: false } },
    plotOptions: {
      bar: { borderRadius: 8, horizontal: false, columnWidth: '60%', endingShape: 'rounded' }
    },
    dataLabels: { enabled: false },
    stroke: { show: true, width: 2, colors: ['transparent'] },
    yaxis: {
      labels: { style: { colors: '#6B7280', fontSize: '12px' }, formatter: intFormatter },
      min: 0,
      forceNiceScale: true
    },
    grid: { borderColor: '#E5E7EB', strokeDashArray: 4 },
    fill: { opacity: 1 }
  };

  function renderChart(selector, options) {
    const el = document.querySelector(selector);
    if (el) new ApexCharts(el, options).render();
    else console.error('Elemento ' + selector + ' não encontrado');
  }

  // Vendas por dia
  renderChart('#purchasesChart', {
    ...areaDefaults,
    series: [{ name: 'Vendas', data: dailyPurchasesCounts }],
    colors: ['#5E72E4']
  });

  // Downloads por dia
  renderChart('#downloadsChart', {
    ...areaDefaults,
    series: [{ name: 'Downloads', data: dailyDownloadsCounts }],
    colors: ['#F56565']
  });

  // Top ebooks enviados
  renderChart('#topEbooksChart', {
    ...barDefaults,
    series: [{ name: 'Total de envios', data: topEbooksCounts }],
    colors: ['#3B82F6'],
    xaxis: {
      categories: topEbooksLabels,
      labels: { style: { colors: '#6B7280', fontSize: '12px' }, rotate: -45, rotateAlways: false }
    },
    tooltip: { theme: 'light', y: { formatter: val => val + ' envios' } }
  });

  // Top ebooks baixados
  renderChart('#topDownloadedEbooksChart', {
    ...barDefaults,
    series: [{ name: 'Total de downloads', data: topDownloadedEbooksCounts }],
    colors: ['#10B981'],
    xaxis: {
      categories: topDownloadedEbooksLabels,
      labels: { style: { colors: '#6B7280', fontSize: '12px' }, rotate: -45, rotateAlways: false }
    },
    tooltip: { theme: 'light', y: { formatter: val => val + ' downloads' } }
  });
})();
